package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lancer/log/internal/telemetry"
)

type SystemStatusHandler struct {
	provider telemetry.Provider
}

func NewSystemStatusHandler(provider telemetry.Provider) *SystemStatusHandler {
	return &SystemStatusHandler{provider: provider}
}

func (h *SystemStatusHandler) Show(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	snapshot, err := h.provider.Snapshot(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, struct {
			Online bool   `json:"online"`
			Error  string `json:"error"`
		}{
			Online: false,
			Error:  "telemetry unavailable",
		})
		return
	}
	c.JSON(http.StatusOK, snapshot)
}
