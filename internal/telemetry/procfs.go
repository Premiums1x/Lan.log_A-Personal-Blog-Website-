package telemetry

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type cpuTimes struct {
	Total uint64
	Idle  uint64
}

type procFSProvider struct {
	root string
	ttl  time.Duration

	mu          sync.Mutex
	cached      Snapshot
	previousCPU cpuTimes
	hasPrevious bool
}

// NewProcFSProvider returns a provider that reads aggregate Linux procfs data.
func NewProcFSProvider(root string, ttl time.Duration) Provider {
	return &procFSProvider{root: root, ttl: ttl}
}

func (p *procFSProvider) Snapshot(ctx context.Context) (Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return Snapshot{}, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return Snapshot{}, err
	}
	now := time.Now().UTC()
	if p.ttl > 0 && !p.cached.SampledAt.IsZero() && now.Sub(p.cached.SampledAt) < p.ttl {
		return p.cached, nil
	}

	stat, err := p.read(ctx, "stat")
	if err != nil {
		return Snapshot{}, err
	}
	meminfo, err := p.read(ctx, "meminfo")
	if err != nil {
		return Snapshot{}, err
	}
	cpuinfo, err := p.read(ctx, "cpuinfo")
	if err != nil {
		return Snapshot{}, err
	}
	uptime, err := p.read(ctx, "uptime")
	if err != nil {
		return Snapshot{}, err
	}

	currentCPU, err := parseCPUStat(string(stat))
	if err != nil {
		return Snapshot{}, err
	}
	memoryTotal, memoryUsed, memoryPercent, err := parseMeminfo(string(meminfo))
	if err != nil {
		return Snapshot{}, err
	}
	frequency, err := parseCPUFrequency(string(cpuinfo))
	if err != nil {
		return Snapshot{}, err
	}
	uptimeSeconds, err := parseUptime(string(uptime))
	if err != nil {
		return Snapshot{}, err
	}

	cpuUsage := 0.0
	if p.hasPrevious {
		cpuUsage = cpuPercent(p.previousCPU, currentCPU)
	}
	snapshot := Snapshot{
		Online:           true,
		CPUPercent:       clampPercent(cpuUsage),
		CPUFrequencyMHz:  frequency,
		MemoryUsedBytes:  memoryUsed,
		MemoryTotalBytes: memoryTotal,
		MemoryPercent:    clampPercent(memoryPercent),
		UptimeSeconds:    uptimeSeconds,
		SampledAt:        time.Now().UTC(),
	}

	p.previousCPU = currentCPU
	p.hasPrevious = true
	p.cached = snapshot
	return snapshot, nil
}

func (p *procFSProvider) read(ctx context.Context, name string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	b, err := os.ReadFile(filepath.Join(p.root, name))
	if err != nil {
		return nil, fmt.Errorf("read procfs %s: unavailable", name)
	}
	return b, nil
}

func parseCPUStat(contents string) (cpuTimes, error) {
	scanner := bufio.NewScanner(strings.NewReader(contents))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 || fields[0] != "cpu" {
			continue
		}
		if len(fields) < 5 {
			return cpuTimes{}, errors.New("procfs stat aggregate CPU line has too few counters")
		}

		counterCount := len(fields) - 1
		if counterCount > 8 {
			counterCount = 8
		}
		counters := make([]uint64, counterCount)
		var total uint64
		for i := 0; i < counterCount; i++ {
			counter, err := strconv.ParseUint(fields[i+1], 10, 64)
			if err != nil {
				return cpuTimes{}, fmt.Errorf("parse procfs CPU counter: %w", err)
			}
			if math.MaxUint64-total < counter {
				return cpuTimes{}, errors.New("procfs CPU counters overflow")
			}
			counters[i] = counter
			total += counter
		}

		idle := counters[3]
		if len(counters) > 4 {
			if math.MaxUint64-idle < counters[4] {
				return cpuTimes{}, errors.New("procfs idle CPU counters overflow")
			}
			idle += counters[4]
		}
		return cpuTimes{Total: total, Idle: idle}, nil
	}
	if err := scanner.Err(); err != nil {
		return cpuTimes{}, fmt.Errorf("scan procfs stat: %w", err)
	}
	return cpuTimes{}, errors.New("procfs stat has no aggregate CPU line")
}

func cpuPercent(previous, current cpuTimes) float64 {
	if current.Total <= previous.Total || current.Idle < previous.Idle {
		return 0
	}
	totalDelta := current.Total - previous.Total
	idleDelta := current.Idle - previous.Idle
	if idleDelta >= totalDelta {
		return 0
	}
	return clampPercent(float64(totalDelta-idleDelta) / float64(totalDelta) * 100)
}

func parseMeminfo(contents string) (total, used uint64, percent float64, err error) {
	values := make(map[string]uint64, 2)
	scanner := bufio.NewScanner(strings.NewReader(contents))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSuffix(fields[0], ":")
		if key != "MemTotal" && key != "MemAvailable" {
			continue
		}
		kilobytes, parseErr := strconv.ParseUint(fields[1], 10, 64)
		if parseErr != nil {
			return 0, 0, 0, fmt.Errorf("parse procfs %s: %w", key, parseErr)
		}
		if kilobytes > math.MaxUint64/1024 {
			return 0, 0, 0, fmt.Errorf("procfs %s overflows bytes", key)
		}
		values[key] = kilobytes * 1024
	}
	if scanErr := scanner.Err(); scanErr != nil {
		return 0, 0, 0, fmt.Errorf("scan procfs meminfo: %w", scanErr)
	}

	total, hasTotal := values["MemTotal"]
	available, hasAvailable := values["MemAvailable"]
	if !hasTotal || total == 0 {
		return 0, 0, 0, errors.New("procfs meminfo has no positive MemTotal")
	}
	if !hasAvailable {
		return 0, 0, 0, errors.New("procfs meminfo has no MemAvailable")
	}
	if available >= total {
		return total, 0, 0, nil
	}
	used = total - available
	return total, used, clampPercent(float64(used) / float64(total) * 100), nil
}

func parseCPUFrequency(contents string) (float64, error) {
	var total float64
	var count uint64
	scanner := bufio.NewScanner(strings.NewReader(contents))
	for scanner.Scan() {
		key, value, found := strings.Cut(scanner.Text(), ":")
		if !found || strings.TrimSpace(key) != "cpu MHz" {
			continue
		}
		frequency, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil || math.IsNaN(frequency) || math.IsInf(frequency, 0) || frequency < 0 {
			return 0, errors.New("procfs cpuinfo has invalid cpu MHz")
		}
		total += frequency
		if math.IsNaN(total) || math.IsInf(total, 0) {
			return 0, errors.New("procfs cpuinfo cpu MHz total is not finite")
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("scan procfs cpuinfo: %w", err)
	}
	if count == 0 {
		return 0, errors.New("procfs cpuinfo has no cpu MHz entries")
	}
	average := total / float64(count)
	if math.IsNaN(average) || math.IsInf(average, 0) {
		return 0, errors.New("procfs cpuinfo average cpu MHz is not finite")
	}
	return average, nil
}

func parseUptime(contents string) (uint64, error) {
	fields := strings.Fields(contents)
	if len(fields) == 0 {
		return 0, errors.New("procfs uptime is empty")
	}
	uptime, err := strconv.ParseFloat(fields[0], 64)
	if err != nil || math.IsNaN(uptime) || math.IsInf(uptime, 0) || uptime < 0 || uptime > math.MaxUint64 {
		return 0, errors.New("procfs uptime is invalid")
	}
	return uint64(uptime), nil
}

func clampPercent(value float64) float64 {
	if math.IsNaN(value) || value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}
