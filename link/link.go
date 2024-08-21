package link

import (
	"strings"

	"golang.org/x/net/html"
)

type Link struct {
	Href string
	Text string
}

// Extract Anchor links from HTML Nodes
func Extract(n *html.Node, links *[]Link, link *Link) {
	if link.Href != "" && n.Type == html.TextNode {
		// trim text, add text to link.Text if it is not empty
		text := strings.TrimSpace(n.Data)
		if text != "" {
			link.Text += text
		}
	}

	if n.Type == html.ElementNode {
		// check for anchor tag with href attribute
		if n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					link.Href = attr.Val
					Extract(n.FirstChild, links, link)
					*links = append(*links, *link)
					link.Href = ""
					link.Text = ""
				}
			}
		}
	}

	if n.FirstChild != nil {
		Extract(n.FirstChild, links, link)
	}

	if n.NextSibling != nil {
		Extract(n.NextSibling, links, link)
	}
}
