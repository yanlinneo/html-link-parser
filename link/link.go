package link

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

type Link struct {
	Href string
	Text string
}

func ExtractLinks(n *html.Node, links *[]Link, link *Link) {
	//fmt.Println("Node:", n.Data, n.Type)

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
					fmt.Println("Node:", n.Data, attr.Key, attr.Val)
					link.Href = attr.Val
					ExtractLinks(n.FirstChild, links, link)
					*links = append(*links, *link)
					link.Href = ""
					link.Text = ""
				}
			}
		}
	}

	if n.FirstChild != nil {
		ExtractLinks(n.FirstChild, links, link)
	}

	if n.NextSibling != nil {
		ExtractLinks(n.NextSibling, links, link)
	}
}
