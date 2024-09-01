package main

import (
	"html-link-parser/models"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

// test with an empty HTML node and ensure empty links are not included
func TestExtract_EmptyLinks(t *testing.T) {
	emptyNode := html.Node{}
	sourceUrl := "https://github.com/yanlinneo"
	got := Extract(&emptyNode, sourceUrl)

	if len(got) > 0 {
		t.Fatalf("Extract() = %v; want links to be empty", got)
	}
}

func TestExtract_ExcludeCommentedText(t *testing.T) {
	htmlString := `<html><body><a href="/dog-cat">dog cat <!-- commented text SHOULD NOT be included! --></a></body></html>`
	sourceUrl := "https://github.com/yanlinneo"
	node, _ := html.Parse(strings.NewReader(htmlString))
	got := Extract(node, sourceUrl)

	if len(got) > 0 {
		for _, l := range got {
			if strings.Contains(l.Text, "commented text SHOULD NOT be included!") {
				t.Fatalf(`Extract() = Commented Text found, commented text should NOT be included.`)
			}
		}
	}
}

func TestExtract_IncludeInnerText(t *testing.T) {
	htmlString := `<html><body><a href="/about-us">Hello! <span>Welcome to about us page!</span></a></body></html>`
	sourceUrl := "https://github.com/yanlinneo"
	node, _ := html.Parse(strings.NewReader(htmlString))
	got := Extract(node, sourceUrl)

	want := []models.Link{
		{Href: "/about-us", Text: "Hello!, Welcome to about us page!", SourceUrl: "https://github.com/yanlinneo"},
	}

	if len(got) != len(want) {
		t.Fatalf("Extract() = %v; want -%v-", got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Extract()[%d] = %v; want %v", i, got[i], want[i])
		}
	}
}
