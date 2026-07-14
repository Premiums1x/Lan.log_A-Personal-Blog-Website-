package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lancer/log/internal/telemetry"
)

type fakeTelemetry struct {
	snapshot telemetry.Snapshot
	err      error
}

func (f fakeTelemetry) Snapshot(context.Context) (telemetry.Snapshot, error) {
	return f.snapshot, f.err
}

func TestSystemStatusReturnsPublicSnapshot(t *testing.T) {
	g := gin.New()
	h := NewSystemStatusHandler(fakeTelemetry{
		snapshot: telemetry.Snapshot{Online: true, CPUPercent: 24.5},
	})
	g.GET("/api/system-status", h.Show)

	r := httptest.NewRequest(http.MethodGet, "/api/system-status", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)

	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"cpu_percent":24.5`) {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if got := w.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}
}

func TestSystemStatusReturnsGenericUnavailableResponse(t *testing.T) {
	const privateDetail = `read procfs stat: open C:\private\proc\stat: access denied`
	g := gin.New()
	h := NewSystemStatusHandler(fakeTelemetry{err: errors.New(privateDetail)})
	g.GET("/api/system-status", h.Show)

	r := httptest.NewRequest(http.MethodGet, "/api/system-status", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
	if got, want := w.Body.String(), `{"online":false,"error":"telemetry unavailable"}`; strings.TrimSpace(got) != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
	if strings.Contains(w.Body.String(), privateDetail) {
		t.Fatalf("body leaked provider error: %s", w.Body.String())
	}
	if got := w.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}
}
