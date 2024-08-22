package main

import (
	"fmt"
	pkglink "html-link-parser/link"
	"os"

	"golang.org/x/net/html"
)

func main() {
	var htmlFile = "ex5.html"
	links, err := parse(htmlFile)

	if err != nil {
		fmt.Println(err)
	}

	for _, l := range links {
		fmt.Println(l.Href, "\t", l.Text)
	}
}

func parse(htmlFile string) ([]pkglink.Link, error) {
	file, err := os.Open(htmlFile)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	node, err := html.Parse(file)

	if err != nil {
		return nil, err
	}

	links := pkglink.Extract(node)

	return links, nil
}
