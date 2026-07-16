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
	assets := []string{"niko-user.jpg", "verstappen-user.jpg", "leclerc-user.jpg", "messi-user.jpg", "curry-new-user.jpg", "gem-user.jpg", "f1-deck-user.jpg", "nba-deck-user.jpg", "cs2-deck-user.jpg", "jay-deck-user.jpg"}
	credits := readInterestMediaSource(t, "static/media/interests/CREDITS.md")
	for _, asset := range assets {
		if _, err := os.Stat("static/media/interests/" + asset); err != nil {
			t.Errorf("interest asset %s missing: %v", asset, err)
		}
		requireInterestMediaStrings(t, credits, "media credits", asset)
	}
	requireInterestMediaStrings(t, credits, "media credits", "User-provided", "No remote image is hotlinked")
}
