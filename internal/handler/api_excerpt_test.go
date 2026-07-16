package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lancer/log/internal/config"
	"github.com/lancer/log/internal/model"
)

func TestOptionalExcerptLeavesBlankInputBlank(t *testing.T) {
	if got := optionalExcerpt("   \n\t"); got != "" {
		t.Fatalf("optionalExcerpt() = %q, want empty string", got)
	}
}

func TestOptionalExcerptKeepsManualCopy(t *testing.T) {
	if got := optionalExcerpt("  手写摘要  "); got != "手写摘要" {
		t.Fatalf("optionalExcerpt() = %q, want trimmed manual excerpt", got)
	}
}

func TestNormalizeExcerptState(t *testing.T) {
	tests := []struct {
		name, excerpt, source, wantExcerpt, wantSource string
	}{
		{name: "blank is explicitly empty", excerpt: "  ", source: "ai", wantExcerpt: "", wantSource: "empty"},
		{name: "ai copy keeps provenance", excerpt: "  AI 摘要  ", source: "ai", wantExcerpt: "AI 摘要", wantSource: "ai"},
		{name: "unknown source becomes manual", excerpt: "手写摘要", source: "", wantExcerpt: "手写摘要", wantSource: "manual"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotExcerpt, gotSource := normalizeExcerptState(tt.excerpt, tt.source)
			if gotExcerpt != tt.wantExcerpt || gotSource != tt.wantSource {
				t.Fatalf("normalizeExcerptState() = (%q, %q), want (%q, %q)", gotExcerpt, gotSource, tt.wantExcerpt, tt.wantSource)
			}
		})
	}
}

func TestBodyContentHashUsesSHA256(t *testing.T) {
	const want = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if got := bodyContentHash("hello"); got != want {
		t.Fatalf("bodyContentHash() = %q, want %q", got, want)
	}
}

func TestExcerptStaleTracksReviewedBody(t *testing.T) {
	reviewed := bodyContentHash("原正文")
	if excerptIsStale(reviewed, "原正文") {
		t.Fatal("matching reviewed body must not be stale")
	}
	if !excerptIsStale(reviewed, "修改后的正文") {
		t.Fatal("changed body must be stale")
	}
	if !excerptIsStale("", "原正文") {
		t.Fatal("missing review hash must be stale")
	}
}

func TestReviewedExcerptHashPreservesOrRefreshesState(t *testing.T) {
	const existing = "old-hash"
	if got := reviewedExcerptHash(existing, "旧摘要", "旧摘要", "正文已改", false); got != existing {
		t.Fatalf("unreviewed draft hash = %q, want preserved %q", got, existing)
	}
	if got := reviewedExcerptHash(existing, "旧摘要", "旧摘要", "正文已改", true); got != bodyContentHash("正文已改") {
		t.Fatalf("explicit review hash = %q", got)
	}
	if got := reviewedExcerptHash(existing, "旧摘要", "手写新摘要", "正文已改", false); got != bodyContentHash("正文已改") {
		t.Fatalf("manual excerpt edit hash = %q", got)
	}
}

func TestDecorateExcerptStateMarksStalePost(t *testing.T) {
	p := model.Post{BodyMD: "正文", ExcerptReviewedBodyHash: bodyContentHash("正文")}
	if got := decorateExcerptState(p); got.ExcerptStale {
		t.Fatal("matching body must not be stale")
	}
	p.BodyMD = "正文已修改"
	if got := decorateExcerptState(p); !got.ExcerptStale {
		t.Fatal("changed body must be stale")
	}
}

func TestGenerateExcerptHandler(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"生成摘要"}}]}`))
	}))
	defer upstream.Close()

	gin.SetMode(gin.TestMode)
	h := NewAPIHandler(nil, nil, config.Config{LLM: config.LLMConfig{
		APIURL: upstream.URL, APIKey: "key", Model: "model", TimeoutSeconds: 2,
	}})
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/ai/excerpt", strings.NewReader(`{"title":"标题","body_md":"正文"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.GenerateExcerpt(c)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "生成摘要") {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestGenerateExcerptHandlerRequiresConfiguration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAPIHandler(nil, nil, config.Config{})
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/ai/excerpt", strings.NewReader(`{"title":"标题","body_md":"正文"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.GenerateExcerpt(c)
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}
