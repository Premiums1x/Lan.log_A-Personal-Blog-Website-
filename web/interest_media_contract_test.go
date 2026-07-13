package web

import (
	"os"
	"strings"
	"testing"
)

func readInterestMediaSource(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

func requireInterestMediaStrings(t *testing.T, source, label string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(source, want) {
			t.Errorf("%s missing %q", label, want)
		}
	}
}

func TestInterestImagesAreLocalAndCredited(t *testing.T) {
	assets := []string{"niko-2022.webp", "verstappen-2018.webp", "leclerc.webp", "messi.webp"}
	credits := readInterestMediaSource(t, "static/media/interests/CREDITS.md")
	for _, asset := range assets {
		if _, err := os.Stat("static/media/interests/" + asset); err != nil {
			t.Errorf("interest asset %s missing: %v", asset, err)
		}
		requireInterestMediaStrings(t, credits, "media credits", asset)
	}
	requireInterestMediaStrings(t, credits, "media credits", "CC BY 4.0", "CC0 1.0", "CC BY-SA 4.0", "CC BY 2.0")
}
