package main

import (
	"strings"
	"testing"
)

func TestPostDocumentIncludesHugoFrontMatter(t *testing.T) {
	document := string(postDocument(articleMetadata{
		Title:       "A title",
		Description: "A description",
		Tags:        []string{"go", "testing"},
	}, "2025-09-07", []byte("Post body\n")))

	for _, expected := range []string{
		"title: \"A title\"",
		"description: \"A description\"",
		"date: 2025-09-07",
		"tags: [\"go\",\"testing\"]",
		"---\n\nPost body",
	} {
		if !strings.Contains(document, expected) {
			t.Errorf("document missing %q:\n%s", expected, document)
		}
	}
}

func TestTitleFromMarkdown(t *testing.T) {
	title, body := titleFromMarkdown([]byte("# Project title\n\nDescription\n"), "fallback")
	if title != "Project title" {
		t.Fatalf("title = %q, want Project title", title)
	}
	if got, want := string(body), "Description\n"; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}
