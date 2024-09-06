package main

import (
	"bytes"
	"errors"
	"fmt"
	"html-link-parser/models"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

var (
	lock         sync.RWMutex
	existingHref = map[string]int{}
)

// Check if Href exists in existingHref map
func checkHref(href string) bool {
	lock.RLock()
	defer lock.RUnlock()

	_, ok := existingHref[href]
	return ok
}

// Add a Href in existingHref map
func saveHref(href string) {
	lock.Lock()
	defer lock.Unlock()

	existingHref[href] = 1
}

func Validate(urlString *string) error {
	parsedUrl, parsedUrlErr := url.Parse(*urlString)

	if parsedUrlErr != nil {
		return parsedUrlErr
	}

	if parsedUrl.Scheme == "" {
		return errors.New("URL should start with https:// or http://")
	}

	if parsedUrl.Host == "" {
		return errors.New("URL is missing a host (e.g. example.com)")
	}

	return nil
}

func getSourceUrl(baseUrl, href string) string {
	// only relative path will have a base url
	if baseUrl != "" {
		return baseUrl + href
	}

	// this should be a full link
	return href
}

func getBaseUrl(href string) string {
	if strings.HasPrefix(href, "/") {
		return baseUrl
	}

	return ""
}

// Extract Anchor links from HTML Nodes
func Extract(n *html.Node, sourceUrl string) []models.Link {
	var links []models.Link
	var text []string

	var getText func(tn *html.Node)
	getText = func(tn *html.Node) {
		if tn == nil {
			return
		}

		if tn.Type == html.TextNode {
			// remove leading and trailing whitespace
			trimText := strings.TrimSpace(tn.Data)

			// replace all whitespace such as \t, \n into a single space
			re := regexp.MustCompile(`\s+`)
			cleanedText := re.ReplaceAllString(trimText, " ")

			if cleanedText != "" {
				text = append(text, cleanedText)
			}
		}

		if tn.FirstChild != nil {
			getText(tn.FirstChild)
		}

		if tn.NextSibling != nil {
			getText(tn.NextSibling)
		}
	}

	var traverse func(n *html.Node)
	traverse = func(n *html.Node) {
		if n == nil {
			return
		}

		// we are only interested in a href FIRST!
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					// we will only be interested in other elements if it is inside a href

					// do not save fragment identifiers, we are only interested in the different links/paths of a site
					if strings.HasPrefix(attr.Val, "#") {
						return
					}

					if exists := checkHref(attr.Val); !exists {
						saveHref(attr.Val)
						getText(n.FirstChild)
						links = append(links, models.Link{Href: attr.Val, Text: strings.Join(text, ", "), SourceUrl: sourceUrl, BaseUrl: baseUrl})
						text = nil

						traverse(n.NextSibling)
						return
					}
				}
			}
		}

		// first child
		if n.FirstChild != nil {
			traverse(n.FirstChild)
		}
		// next sibling
		if n.NextSibling != nil {
			traverse(n.NextSibling)
		}
	}

	traverse(n)
	return links
}

// Runs process function concurrently
func concurrentProcess(pendingLinks []models.Link, linkRepo models.LinkRepository) {
	// Buffered channel to limit concurrent goroutines
	semaphore := make(chan struct{}, 10)
	var wg sync.WaitGroup

	for _, pl := range pendingLinks {
		wg.Add(1) // Increment WaitGroup counter

		go func(link models.Link) {
			defer wg.Done() // Decrement WaitGroup counter when the goroutine completes

			semaphore <- struct{}{} // Acquire a slot in the semaphore
			defer func() {
				<-semaphore // Release the slot in the semaphore
			}()

			process(&link, linkRepo)
			//time.Sleep(1 * time.Second)
		}(pl) // Pass the variable l to avoid closure capture issues
	}

	wg.Wait() // Wait for all goroutines to finish
}

func process(pendingLink *models.Link, linkRepo models.LinkRepository) {
	// get full source url
	sourceUrl := getSourceUrl(pendingLink.BaseUrl, pendingLink.Href)

	// call the url
	body, statusCode, statusMessage, respErr := call(sourceUrl)

	// set the status code and message
	if statusCode >= 0 {
		pendingLink.StatusCode.Int32 = int32(statusCode)
		pendingLink.StatusMessage.String = statusMessage
		_, dbErr := linkRepo.UpdateStatus(*pendingLink)

		if dbErr != nil {
			slog.Error("Database UpdateStatus", "error", dbErr) // will continue
		}

		relativePathsAccessed++
	}

	// print the respone error
	if respErr != nil {
		slog.Error("Response", "error", respErr)
		return
	}

	// parse the response body
	node, parseErr := parse(body)
	if parseErr != nil {
		slog.Error("Parsed", "error", parseErr)
		return
	}

	// extract nodes into links
	extractedLinks := Extract(node, sourceUrl)

	// save the extracted links
	_, dbErr := linkRepo.AddBulk(extractedLinks)
	if dbErr != nil {
		slog.Info("Database AddBulk", "error", dbErr)
	}
}

// Accessing a URL via API
func call(url string) ([]byte, int, string, error) {
	//now := time.Now()

	client := &http.Client{
		Timeout: 10 * time.Second, // Set timeout = 10s
	}

	resp, respErr := client.Get(url)
	if respErr != nil {
		return nil, 0, respErr.Error(), respErr
	}

	// Close response body
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// if time.Since(now) > 1*time.Second {
		// 	return nil, resp.StatusCode, resp.Status, fmt.Errorf("LONG HTTP request failed: url:%s time:%s, took: %v", url, resp.Status, time.Since(now))
		// }
		return nil, resp.StatusCode, resp.Status, nil
		//return nil, resp.StatusCode, resp.Status, fmt.Errorf("HTTP request failed: url:%s status:%s, took: %v", url, resp.Status, time.Since(now))
	}

	if baseUrl == "" {
		baseUrl = fmt.Sprintf("%s://%s", resp.Request.URL.Scheme, resp.Request.URL.Hostname())
		baseUrlHost = strings.Replace(resp.Request.URL.Hostname(), ".", "_", -1)
	}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, resp.StatusCode, resp.Status, readErr
	}

	return body, resp.StatusCode, resp.Status, nil
}

// Converting Response Body into a Node and calls Extract function
func parse(body []byte) (*html.Node, error) {
	node, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	return node, nil
}
