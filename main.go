package main

import (
	"bytes"
	"fmt"
	pkglink "html-link-parser/link"
	"html-link-parser/repository"
	"io"
	"net/http"
	"time"

	"golang.org/x/net/html"
)

func main() {
	start := time.Now()
	var sourceUrl = "https://google.com.sg"

	respBody, respErr := call(sourceUrl)
	if respErr != nil {
		fmt.Println(respErr)
		return
	}

	dbErr := repository.InitDB()
	if dbErr != nil {
		fmt.Println(dbErr)
		return
	}

	links, parseErr := parse(respBody)
	if parseErr != nil {
		fmt.Println(parseErr)
		return
	}

	saveLinks(links, sourceUrl)

	fmt.Println("Total Duration: ", time.Since(start))
}

func call(url string) ([]byte, error) {
	resp, respErr := http.Post(url, "application/html", nil)
	if respErr != nil {
		return nil, respErr
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP request failed with :%s", resp.Status)
	}

	// Close response body
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}

	return body, nil
}

func parse(body []byte) ([]pkglink.Link, error) {
	node, err := html.Parse(bytes.NewReader(body))

	if err != nil {
		return nil, err
	}

	links := pkglink.Extract(node)

	return links, nil
}

func saveLinks(links []pkglink.Link, sourceUrl string) {
	// save in database
	var save, notSave int
	for _, l := range links {
		l.SourceUrl = sourceUrl
		_, err := repository.AddLink(l)
		if err != nil {
			notSave++
		} else {
			save++
		}
	}

	fmt.Println("Save:", save, "records")
	fmt.Println("Did not save:", notSave, "records")
}
