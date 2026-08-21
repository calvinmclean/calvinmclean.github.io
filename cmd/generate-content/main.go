// Command generate-content converts the article-sync source directories into
// temporary Hugo page bundles. The generated tree is deliberately untracked:
// articles/ remains the single source of truth for both DEV and this site.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	articlesDir = "articles"
	projectsDir = "projects"
	outputDir   = ".hugo-content"
)

type articleMetadata struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

func main() {
	if err := generate(); err != nil {
		fmt.Fprintln(os.Stderr, "generate Hugo content:", err)
		os.Exit(1)
	}
}

func generate() error {
	if err := os.RemoveAll(outputDir); err != nil {
		return err
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}

	if err := writeHome(); err != nil {
		return err
	}
	if err := generateArticles(); err != nil {
		return err
	}
	return generateProjects()
}

func writeHome() error {
	const home = `---
title: "Calvin McLean"
---

I build open-source tools and write about Go, developer productivity, and the ideas I am currently exploring.
`
	return os.WriteFile(filepath.Join(outputDir, "_index.md"), []byte(home), 0o644)
}

func generateArticles() error {
	entries, err := os.ReadDir(articlesDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		slug := entry.Name()
		source := filepath.Join(articlesDir, slug)
		metadata, body, err := readArticle(source)
		if err != nil {
			return fmt.Errorf("%s: %w", source, err)
		}
		if metadata.Title == "" {
			return fmt.Errorf("%s: article.json has no title", source)
		}

		date, err := firstCommitDate(filepath.Join(source, "article.md"))
		if err != nil {
			return err
		}
		destination := filepath.Join(outputDir, "posts", slug)
		if err := os.MkdirAll(destination, 0o755); err != nil {
			return err
		}
		if err := copyMedia(source, destination); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(destination, "index.md"), postDocument(metadata, date, body), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func readArticle(source string) (articleMetadata, []byte, error) {
	var metadata articleMetadata
	rawMetadata, err := os.ReadFile(filepath.Join(source, "article.json"))
	if err != nil {
		return metadata, nil, err
	}
	if err := json.Unmarshal(rawMetadata, &metadata); err != nil {
		return metadata, nil, err
	}
	body, err := os.ReadFile(filepath.Join(source, "article.md"))
	return metadata, body, err
}

func postDocument(metadata articleMetadata, date string, body []byte) []byte {
	var document bytes.Buffer
	document.WriteString("---\n")
	document.WriteString("title: ")
	document.WriteString(strconv.Quote(metadata.Title))
	document.WriteString("\n")
	if metadata.Description != "" {
		document.WriteString("description: ")
		document.WriteString(strconv.Quote(metadata.Description))
		document.WriteString("\n")
	}
	document.WriteString("date: ")
	document.WriteString(date)
	document.WriteString("\n")
	if len(metadata.Tags) > 0 {
		tags, _ := json.Marshal(metadata.Tags)
		document.WriteString("tags: ")
		document.Write(tags)
		document.WriteString("\n")
	}
	document.WriteString("---\n\n")
	document.Write(body)
	return document.Bytes()
}

func firstCommitDate(path string) (string, error) {
	command := exec.Command("git", "log", "--diff-filter=A", "--format=%as", "--", path)
	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("read first commit date for %s: %w", path, err)
	}
	lines := strings.Fields(string(output))
	if len(lines) == 0 {
		return "", fmt.Errorf("no commit date found for %s", path)
	}
	return lines[len(lines)-1], nil
}

func copyMedia(source, destination string) error {
	entries, err := os.ReadDir(source)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "article.md" || entry.Name() == "article.json" {
			continue
		}
		input, err := os.Open(filepath.Join(source, entry.Name()))
		if err != nil {
			return err
		}
		output, err := os.Create(filepath.Join(destination, entry.Name()))
		if err != nil {
			_ = input.Close()
			return err
		}
		_, err = io.Copy(output, input)
		closeErr := errors.Join(output.Close(), input.Close())
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

func generateProjects() error {
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		body, err := os.ReadFile(filepath.Join(projectsDir, entry.Name()))
		if err != nil {
			return err
		}
		date, err := firstCommitDate(filepath.Join(projectsDir, entry.Name()))
		if err != nil {
			return err
		}
		title, body := titleFromMarkdown(body, strings.TrimSuffix(entry.Name(), ".md"))
		slug := strings.TrimSuffix(entry.Name(), ".md")
		destination := filepath.Join(outputDir, "projects", slug)
		if err := os.MkdirAll(destination, 0o755); err != nil {
			return err
		}
		document := "---\ntitle: " + strconv.Quote(title) + "\ndate: " + date + "\n---\n\n" + string(body)
		if err := os.WriteFile(filepath.Join(destination, "index.md"), []byte(document), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func titleFromMarkdown(body []byte, fallback string) (string, []byte) {
	text := string(body)
	if !strings.HasPrefix(text, "# ") {
		return fallback, body
	}
	lineEnd := strings.IndexByte(text, '\n')
	if lineEnd == -1 {
		return strings.TrimPrefix(text, "# "), nil
	}
	return strings.TrimSpace(strings.TrimPrefix(text[:lineEnd], "# ")), []byte(strings.TrimLeft(text[lineEnd+1:], "\n"))
}
