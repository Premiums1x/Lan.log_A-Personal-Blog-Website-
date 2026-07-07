package handler

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"html/template"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/lancer/log/internal/markdown"
)

func randomStamp() int64 { return time.Now().UnixNano() }

// commitShort produces a stable 7-char hash from any input (or empty → placeholder).
func commitShort(s string) string {
	if s == "" {
		// deterministic-ish placeholder; real hash set at insert time
		return "0000000"
	}
	if len(s) >= 7 {
		return s[:7]
	}
	return s
}

// firstChar returns the first rune (for the avatar circle).
func firstChar(s string) string {
	if s == "" {
		return "?"
	}
	r, _ := utf8.DecodeRuneInString(s)
	return string(r)
}

// renderInline renders a one-line markdown title to inline HTML (strip leading #).
func renderInline(s string) template.HTML {
	s = strings.TrimSpace(strings.TrimPrefix(s, "#"))
	return template.HTML(markdown.Render(s))
}

// splitAccent wraps the first occurrence of accent inside title with <span class="accent">.
// Also converts literal "\n" or "<br>" stays as-is. We keep <br> in title strings.
func splitAccent(title, accent string) template.HTML {
	if accent == "" || !strings.Contains(title, accent) {
		return template.HTML(title)
	}
	idx := strings.Index(title, accent)
	return template.HTML(title[:idx] + `<span class="accent">` + accent + `</span>` + title[idx+len(accent):])
}

// newCommitHash returns a fresh 7-char hex hash (used when creating a post).
func newCommitHash(slug string) string {
	h := sha1.Sum([]byte(fmt.Sprintf("%s-%d", slug, randomStamp())))
	return hex.EncodeToString(h[:])[:7]
}