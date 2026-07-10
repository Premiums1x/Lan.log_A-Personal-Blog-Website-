package web

import (
	"os"
	"strings"
	"testing"
)

func TestPostPageUsesIndependentReadingPaper(t *testing.T) {
	template, err := os.ReadFile("templates/post.tmpl")
	if err != nil {
		t.Fatalf("read post template: %v", err)
	}
	css, err := os.ReadFile("static/style.css")
	if err != nil {
		t.Fatalf("read stylesheet: %v", err)
	}

	for _, want := range []string{
		`{{if .Post.Excerpt}}`,
		`class="article-paper`,
		`class="art-body`,
	} {
		if !strings.Contains(string(template), want) {
			t.Errorf("post template missing %q", want)
		}
	}

	for _, want := range []string{
		`.article-paper {`,
		`.article-paper::before`,
		`@keyframes article-aura-drift`,
		`@media (max-width: 720px)`,
	} {
		if !strings.Contains(string(css), want) {
			t.Errorf("article paper styling missing %q", want)
		}
	}
}
