package main

import (
	"fmt"
	"html-link-parser/link"
	pkglink "html-link-parser/link"
	"os"

	"golang.org/x/net/html"
)

func main() {
	parseHTML()
}

func parseHTML() {
	file, err := os.Open("ex4.html")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}

	defer file.Close()

	node, err := html.Parse(file)

	if err != nil {
		fmt.Println(err)
	}

	var links []link.Link
	var link link.Link
	pkglink.ExtractLinks(node, &links, &link)

	fmt.Println("let's loop!")
	for _, l := range links {
		fmt.Println(l.Href, "------>", l.Text)
	}
}
