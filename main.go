package main

import (
	"bytes"
	"fmt"
	pkglink "html-link-parser/link"
	"html-link-parser/repository"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

var baseUrl string

func main() {
	start := time.Now()
	var err error

	// start the program with 1 link (sourceLink)
	var sourceLink = pkglink.Link{Href: ""}

	// add sourceLink to pendingLinks
	var pendingLinks []pkglink.Link
	pendingLinks = append(pendingLinks, sourceLink)

	// start db
	dbErr := repository.InitDB()
	if dbErr != nil {
		slog.Error("Failed to initialize DB:", "error", dbErr)
		return
	}

	// pendingLinks will be processed in this loop
	for {
		concurrentProcess(pendingLinks)

		// fetch all relative paths (pendingLinks) from the database
		pendingLinks, err = repository.RelativePaths(baseUrl)
		if err != nil {
			slog.Error("Relative Path Links:", "error", err)
		}

		// if there are no more pendingLinks to process, break
		if len(pendingLinks) == 0 {
			break
		}
	}

	fmt.Println("Total Duration: ", time.Since(start))
}

func concurrentProcess(pendingLinks []pkglink.Link) {
	concurrentStart := time.Now()

	// Buffered channel to limit concurrent goroutines
	semaphore := make(chan struct{}, 2)
	var wg sync.WaitGroup

	for _, pl := range pendingLinks {
		wg.Add(1) // Increment WaitGroup counter

		go func(link pkglink.Link) {
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
	slog.Info("Duration: ", "duration", time.Since(concurrentStart))
}

func process(pendingLink *pkglink.Link) {
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
		_, dbErr := repository.UpdateStatus(*pendingLink)

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

func parse(body []byte) ([]pkglink.Link, error) {
	node, err := html.Parse(bytes.NewReader(body))

	if err != nil {
		return nil, err
	}

	links := pkglink.Extract(node)

	return links, nil
}

func save(links []pkglink.Link, sourceUrl string) {
	// save in database
	var save, notSave int

	for _, l := range links {
		l.SourceUrl = sourceUrl

		// check if Href is a relative path, if yes, we set the base url
		if strings.HasPrefix(l.Href, "/") {
			l.BaseUrl = baseUrl
		}

		// save all the links, to be improvised
		_, err := repository.Add(l)
		if err != nil {
			notSave++
		} else {
			save++
		}
	}

	fmt.Println("Save:", save, "records")
	fmt.Println("Did not save:", notSave, "records")
}
