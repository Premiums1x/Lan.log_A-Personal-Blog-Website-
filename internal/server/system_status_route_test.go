package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lancer/log/internal/handler"
	"github.com/lancer/log/internal/telemetry"
)

type fakeSystemStatusTelemetry struct {
	snapshot telemetry.Snapshot
}

func (f fakeSystemStatusTelemetry) Snapshot(context.Context) (telemetry.Snapshot, error) {
	return f.snapshot, nil
}

func TestSystemStatusRouteIsPublic(t *testing.T) {
	g := gin.New()
	status := handler.NewSystemStatusHandler(fakeSystemStatusTelemetry{
		snapshot: telemetry.Snapshot{Online: true, CPUPercent: 24.5},
	})
	registerSystemStatusRoute(g, status.Show)

	private := g.Group("/api")
	private.Use(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusUnauthorized)
	})
	private.GET("/private", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	privateRequest := httptest.NewRequest(http.MethodGet, "/api/private", nil)
	privateResponse := httptest.NewRecorder()
	g.ServeHTTP(privateResponse, privateRequest)
	if privateResponse.Code != http.StatusUnauthorized {
		t.Fatalf("private status = %d, want %d", privateResponse.Code, http.StatusUnauthorized)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/system-status", nil)
	response := httptest.NewRecorder()
	g.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("system status = %d, want %d; body = %s", response.Code, http.StatusOK, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"cpu_percent":24.5`) {
		t.Fatalf("system status body = %s", response.Body.String())
	}
}
