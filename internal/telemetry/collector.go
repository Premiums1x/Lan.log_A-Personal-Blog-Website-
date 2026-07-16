package telemetry

import (
	"context"
	"time"
)

// Snapshot contains aggregate host health data without host identity details.
type Snapshot struct {
	Online           bool      `json:"online"`
	CPUPercent       float64   `json:"cpu_percent"`
	CPUFrequencyMHz  float64   `json:"cpu_frequency_mhz"`
	MemoryUsedBytes  uint64    `json:"memory_used_bytes"`
	MemoryTotalBytes uint64    `json:"memory_total_bytes"`
	MemoryPercent    float64   `json:"memory_percent"`
	UptimeSeconds    uint64    `json:"uptime_seconds"`
	SampledAt        time.Time `json:"sampled_at"`
}

// Provider supplies privacy-safe aggregate telemetry snapshots.
type Provider interface {
	Snapshot(context.Context) (Snapshot, error)
}
