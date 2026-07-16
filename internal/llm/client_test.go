package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lancer/log/internal/config"
)

func TestGenerateExcerptSendsCompatibleRequest(t *testing.T) {
	var gotModel string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("Authorization = %q", got)
		}
		var body struct {
			Model    string `json:"model"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		gotModel = body.Model
		if len(body.Messages) != 2 || body.Messages[1].Content == "" {
			t.Fatalf("unexpected messages: %#v", body.Messages)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"  这是生成后的摘要。  "}}]}`))
	}))
	defer server.Close()

	client := NewClient(config.LLMConfig{APIURL: server.URL, APIKey: "test-key", Model: "test-model", TimeoutSeconds: 2})
	got, err := client.GenerateExcerpt(context.Background(), "标题", "正文内容")
	if err != nil {
		t.Fatal(err)
	}
	if got != "这是生成后的摘要。" {
		t.Fatalf("excerpt = %q", got)
	}
	if gotModel != "test-model" {
		t.Fatalf("model = %q", gotModel)
	}
}

func TestGenerateExcerptRejectsHTTPFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "upstream failed", http.StatusBadGateway)
	}))
	defer server.Close()

	client := NewClient(config.LLMConfig{APIURL: server.URL, APIKey: "key", Model: "model", TimeoutSeconds: 2})
	if _, err := client.GenerateExcerpt(context.Background(), "标题", "正文"); err == nil {
		t.Fatal("expected upstream error")
	}
}

func TestGenerateExcerptRejectsEmptyContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"  "}}]}`))
	}))
	defer server.Close()

	client := NewClient(config.LLMConfig{APIURL: server.URL, APIKey: "key", Model: "model", TimeoutSeconds: int((2 * time.Second) / time.Second)})
	if _, err := client.GenerateExcerpt(context.Background(), "标题", "正文"); err == nil {
		t.Fatal("expected empty-content error")
	}
}
