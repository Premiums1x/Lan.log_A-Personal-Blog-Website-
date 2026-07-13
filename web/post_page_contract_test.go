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
		`.article-paper .art-trace`,
		`.art-trace .trace-path`,
		`@media (max-width: 720px)`,
	} {
		if !strings.Contains(string(css), want) {
			t.Errorf("article paper styling missing %q", want)
		}
	}
}

func TestPostPageDoesNotFallBackToAuthorCardWithoutHeatmap(t *testing.T) {
	template, err := os.ReadFile("templates/post.tmpl")
	if err != nil {
		t.Fatalf("read post template: %v", err)
	}

	source := string(template)
	if !strings.Contains(source, `{{if .Heatmap}}`) {
		t.Fatal("post template must conditionally render the GitHub heatmap")
	}
	for _, unwanted := range []string{
		`class="author-bio"`,
		`读全部文章`,
	} {
		if strings.Contains(source, unwanted) {
			t.Errorf("post template still contains heatmap fallback %q", unwanted)
		}
	}
}

func TestArticleTraceUsesDeliberateSteppedMotionAndFreezesWhileScrolling(t *testing.T) {
	template, err := os.ReadFile("templates/post.tmpl")
	if err != nil {
		t.Fatalf("read post template: %v", err)
	}
	css, err := os.ReadFile("static/style.css")
	if err != nil {
		t.Fatalf("read stylesheet: %v", err)
	}

	for _, want := range []string{
		`var MOVE_INTERVAL = 120;`,
		`var POSITION_STEP = 36;`,
		`if (scrolling) return;`,
		`window.addEventListener('scroll'`,
	} {
		if !strings.Contains(string(template), want) {
			t.Errorf("article trace motion missing %q", want)
		}
	}
	if !strings.Contains(string(css), `transition: stroke-dashoffset .16s`) {
		t.Error("article trace transition must settle before the next sampled move")
	}
}

func TestArticleTraceOnlyRespondsNearEdgesAndUsesShortestRoute(t *testing.T) {
	template, err := os.ReadFile("templates/post.tmpl")
	if err != nil {
		t.Fatalf("read post template: %v", err)
	}

	for _, want := range []string{
		`var EDGE_ZONE = 80;`,
		`var EDGE_HYSTERESIS = 20;`,
		`function nearestEdge(x, y, W, H)`,
		`if (!edge) {`,
		`function unwrapOffset(off)`,
		`Math.round((currentOffset - off) / peri) * peri`,
	} {
		if !strings.Contains(string(template), want) {
			t.Errorf("article trace edge behavior missing %q", want)
		}
	}
}

func TestAdminEditorReviewsAIExcerptBeforePublishing(t *testing.T) {
	source, err := os.ReadFile("admin/src/pages/PostEdit.tsx")
	if err != nil {
		t.Fatalf("read post editor: %v", err)
	}
	for _, want := range []string{
		`/api/ai/excerpt`,
		`正文已修改，当前摘要可能已经过期`,
		`AI 生成最新摘要`,
		`保留当前摘要`,
		`清空摘要`,
		`excerpt_reviewed`,
		`excerpt_stale`,
	} {
		if !strings.Contains(string(source), want) {
			t.Errorf("post editor AI excerpt workflow missing %q", want)
		}
	}
}
