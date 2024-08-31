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

// Extract Anchor links from HTML Nodes
func Extract(n *html.Node) []models.Link {
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
						links = append(links, models.Link{Href: attr.Val, Text: strings.Join(text, ", ")})
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
func concurrentProcess(pendingLinks []models.Link) {
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

			process(&link)
			time.Sleep(2 * time.Second)
		}(pl) // Pass the variable l to avoid closure capture issues
	}

	wg.Wait() // Wait for all goroutines to finish
}

func process(pendingLink *models.Link) {
	// if there is a base url, base url + href (relative path)
	var url string
	if pendingLink.BaseUrl != "" {
		url = pendingLink.BaseUrl
	}

	url += pendingLink.Href

	// call the url
	body, statusCode, statusMessage, respErr := call(url)

	// set the status code and message
	if statusCode != 0 {
		pendingLink.StatusCode = statusCode
		pendingLink.StatusMessage = statusMessage
		_, dbErr := pendingLink.UpdateStatus()

		if dbErr != nil {
			slog.Error("Database UpdateStatus Error:", "error", dbErr) // will continue
		}
	}

	// print the respone error
	if respErr != nil {
		slog.Error("Response Error:", "error", respErr)
		return
	}

	// parse the response body
	extractedLinks, parseErr := parse(body)
	if parseErr != nil {
		slog.Error("Parsed Error:", "error", parseErr)
		return
	}

	// save the extracted links
	save(extractedLinks, url)
}

// Accessing a URL via API
func call(url string) ([]byte, int, string, error) {
	resp, respErr := http.Get(url)
	if respErr != nil {
		return nil, 0, "", respErr
	}

	// Close response body
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, resp.StatusCode, resp.Status, fmt.Errorf("HTTP request failed with:%s", resp.Status)
	}

	if baseUrl == "" {
		baseUrl = fmt.Sprintf("%s://%s", resp.Request.URL.Scheme, resp.Request.URL.Host)
	}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, resp.StatusCode, resp.Status, readErr
	}

	return body, resp.StatusCode, resp.Status, nil
}

// Converting Response Body into a Node and calls Extract function
func parse(body []byte) ([]models.Link, error) {
	node, err := html.Parse(bytes.NewReader(body))

	if err != nil {
		return nil, err
	}

	links := Extract(node)

	return links, nil
}

// Saving Links to database. To explore if a batch save will be better.
func save(links []models.Link, sourceUrl string) {
	// save in database
	var save, notSave int

	for _, l := range links {
		l.SourceUrl = sourceUrl

		// check if Href is a relative path, if yes, we set the base url
		if strings.HasPrefix(l.Href, "/") {
			l.BaseUrl = baseUrl
		}

		// save all the links, to be improvised
		_, err := l.Add()
		if err != nil {
			slog.Info("Error when adding:", "error", err)
			slog.Info("Links when adding", "Href", l.Href, "Base", l.BaseUrl, "Source", l.SourceUrl, "Text", l.Text)
			notSave++
		} else {
			save++
		}
	}

	slog.Info("Saving into DB:", "Links saved", save, "Links not saved", notSave)
}
