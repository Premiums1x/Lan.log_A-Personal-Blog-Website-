package telemetry

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseMeminfoUsesMemAvailable(t *testing.T) {
	total, used, pct, err := parseMeminfo("MemTotal: 1000 kB\nMemAvailable: 375 kB\n")
	if err != nil || total != 1024000 || used != 640000 || pct != 62.5 {
		t.Fatalf("got total=%d used=%d pct=%v err=%v", total, used, pct, err)
	}
}

func TestParseMeminfoClampsAvailableAboveTotal(t *testing.T) {
	total, used, pct, err := parseMeminfo("MemTotal: 1000 kB\nMemAvailable: 1200 kB\n")
	if err != nil || total != 1024000 || used != 0 || pct != 0 {
		t.Fatalf("got total=%d used=%d pct=%v err=%v", total, used, pct, err)
	}
}

func TestCPUPercentUsesCounterDelta(t *testing.T) {
	previous := cpuTimes{Total: 1000, Idle: 700}
	current := cpuTimes{Total: 1200, Idle: 760}
	if got := cpuPercent(previous, current); got != 70 {
		t.Fatalf("got %v", got)
	}
}

func TestCPUPercentClampsInvalidCounterDelta(t *testing.T) {
	tests := []struct {
		name     string
		previous cpuTimes
		current  cpuTimes
	}{
		{name: "unchanged", previous: cpuTimes{Total: 100, Idle: 20}, current: cpuTimes{Total: 100, Idle: 20}},
		{name: "counter reset", previous: cpuTimes{Total: 100, Idle: 20}, current: cpuTimes{Total: 50, Idle: 10}},
		{name: "idle exceeds total delta", previous: cpuTimes{Total: 100, Idle: 20}, current: cpuTimes{Total: 110, Idle: 40}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cpuPercent(tt.previous, tt.current); got != 0 {
				t.Fatalf("got %v", got)
			}
		})
	}
}

func TestParseCPUFrequencyAveragesAllEntries(t *testing.T) {
	got, err := parseCPUFrequency("processor: 0\ncpu MHz: 2400.000\nprocessor: 1\ncpu MHz: 3600.000\n")
	if err != nil || got != 3000 {
		t.Fatalf("got frequency=%v err=%v", got, err)
	}
}

func TestSnapshotJSONDoesNotExposeHostIdentity(t *testing.T) {
	b, err := json.Marshal(Snapshot{})
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"hostname", "process", "path", "ip", "mount"} {
		if strings.Contains(strings.ToLower(string(b)), forbidden) {
			t.Fatalf("leaked %s in %s", forbidden, b)
		}
	}
}

func TestProcFSProviderReadsAggregateSnapshotAndCachesWithinTTL(t *testing.T) {
	root := t.TempDir()
	writeProcFixture(t, root, "stat", "cpu  100 20 30 700 50 10 5 0 0 0\n")
	writeProcFixture(t, root, "meminfo", "MemTotal: 1000 kB\nMemAvailable: 375 kB\n")
	writeProcFixture(t, root, "cpuinfo", "processor: 0\ncpu MHz: 2400.000\nprocessor: 1\ncpu MHz: 3600.000\n")
	writeProcFixture(t, root, "uptime", "123.45 67.89\n")

	provider := NewProcFSProvider(root, 5*time.Second)
	first, err := provider.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !first.Online || first.CPUPercent != 0 || first.CPUFrequencyMHz != 3000 ||
		first.MemoryUsedBytes != 640000 || first.MemoryTotalBytes != 1024000 ||
		first.MemoryPercent != 62.5 || first.UptimeSeconds != 123 || first.SampledAt.IsZero() {
		t.Fatalf("unexpected snapshot: %+v", first)
	}

	writeProcFixture(t, root, "uptime", "999.00 67.89\n")
	second, err := provider.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if second.SampledAt != first.SampledAt {
		t.Fatalf("cache miss: first=%v second=%v", first.SampledAt, second.SampledAt)
	}
	if second.UptimeSeconds != first.UptimeSeconds {
		t.Fatalf("cached snapshot changed: first=%d second=%d", first.UptimeSeconds, second.UptimeSeconds)
	}
}

func TestProcFSProviderUsesCounterDeltaAfterCacheExpires(t *testing.T) {
	root := t.TempDir()
	writeProcFixture(t, root, "stat", "cpu  100 0 200 700\n")
	writeProcFixture(t, root, "meminfo", "MemTotal: 1000 kB\nMemAvailable: 375 kB\n")
	writeProcFixture(t, root, "cpuinfo", "cpu MHz: 2500\n")
	writeProcFixture(t, root, "uptime", "10.0 0.0\n")

	provider := NewProcFSProvider(root, 0)
	first, err := provider.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if first.CPUPercent != 0 {
		t.Fatalf("first sample CPU percent = %v", first.CPUPercent)
	}

	writeProcFixture(t, root, "stat", "cpu  170 0 270 760\n")
	second, err := provider.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(second.CPUPercent-70) > 0.000001 {
		t.Fatalf("second sample CPU percent = %v", second.CPUPercent)
	}
}

func TestProcFSProviderHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := NewProcFSProvider(t.TempDir(), time.Second).Snapshot(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got err=%v", err)
	}
}

func TestProcFSProviderErrorDoesNotExposeRootPath(t *testing.T) {
	root := t.TempDir()

	_, err := NewProcFSProvider(root, time.Second).Snapshot(context.Background())
	if err == nil {
		t.Fatal("expected missing fixture error")
	}
	if strings.Contains(err.Error(), root) {
		t.Fatalf("error exposed procfs root: %v", err)
	}
}

func writeProcFixture(t *testing.T, root, name, contents string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, name), []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}
