package link

import (
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

type Link struct {
	ID              int64
	Href            string
	Text            string
	SourceUrl       string
	BaseUrl         string
	CreatedDateTime time.Time
	StatusCode      int
	StatusMessage   string
}

func checkHref(href string) bool {
	lock.RLock()
	defer lock.RUnlock()

	_, ok := existingHref[href]
	return ok
}

func saveHref(href string) {
	lock.Lock()
	defer lock.Unlock()

	existingHref[href] = 1
}

// Extract Anchor links from HTML Nodes
func Extract(n *html.Node) []Link {
	var links []Link
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

					if exists := checkHref(attr.Val); !exists {
						saveHref(attr.Val)
						getText(n.FirstChild)
						links = append(links, Link{Href: attr.Val, Text: strings.Join(text, ", ")})
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
