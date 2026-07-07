package markdown

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	gmhtml "github.com/yuin/goldmark/renderer/html"
)

var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM, extension.Linkify),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	goldmark.WithRendererOptions(gmhtml.WithUnsafe(), gmhtml.WithHardWraps()),
)

// Render converts markdown to HTML. Headings/manual <pre> styling handled in templates/CSS.
func Render(src string) string {
	var buf bytes.Buffer
	if err := md.Convert([]byte(src), &buf); err != nil {
		return src
	}
	return buf.String()
}

// WordCount rough CJK + latin word count.
func WordCount(src string) int {
	n := 0
	inWord := false
	for _, r := range src {
		switch {
		case r > 0x4E00 && r < 0x9FFF, r > 0x3000 && r < 0x303F: // cjk + cjk punctuation
			n++
			inWord = false
		case r == ' ' || r == '\n' || r == '\t' || r == '\r':
			inWord = false
		default:
			if !inWord {
				n++
				inWord = true
			}
		}
	}
	return n
}

// ReadMinutes estimates reading time. CJK ~400/min, latin ~250 wpm.
func ReadMinutes(src string) int {
	w := WordCount(src)
	mins := w / 350
	if mins < 1 {
		mins = 1
	}
	return mins
}

// Excerpt plain-text first paragraph truncated.
func Excerpt(mdSrc string, max int) string {
	s := strings.TrimSpace(mdSrc)
	if i := strings.Index(s, "\n\n"); i > 0 {
		s = s[:i]
	}
	s = strings.TrimSpace(strings.ReplaceAll(s, "#", ""))
	if max > 0 && len([]rune(s)) > max {
		s = string([]rune(s)[:max]) + "…"
	}
	return s
}