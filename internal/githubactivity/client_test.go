package githubactivity

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestClientFetchesAndCachesContributionCalendar(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if got := r.Header.Get("Authorization"); got != "Bearer token" {
			t.Fatalf("Authorization = %q", got)
		}
		_, _ = w.Write([]byte(`{"data":{"user":{"login":"Premiums1x","url":"https://github.com/Premiums1x","contributionsCollection":{"contributionCalendar":{"totalContributions":12,"weeks":[{"contributionDays":[{"date":"2026-07-10","contributionCount":3,"contributionLevel":"FIRST_QUARTILE"}]}]}}}}}`))
	}))
	defer server.Close()

	client := NewClient(Config{Username: "Premiums1x", Token: "token", Endpoint: server.URL, CacheTTL: time.Hour})
	first, err := client.Activity(context.Background())
	if err != nil {
		t.Fatalf("first Activity: %v", err)
	}
	second, err := client.Activity(context.Background())
	if err != nil {
		t.Fatalf("second Activity: %v", err)
	}
	if first.TotalContributions != 12 || first.Weeks[0].Days[0].Count != 3 {
		t.Fatalf("unexpected activity: %#v", first)
	}
	if second.Username != "Premiums1x" || calls.Load() != 1 {
		t.Fatalf("calls = %d, want one cached request", calls.Load())
	}
}

func TestClientSkipsFetchWithoutCredentials(t *testing.T) {
	activity, err := NewClient(Config{}).Activity(context.Background())
	if err != nil || activity != nil {
		t.Fatalf("Activity() = %#v, %v; want nil, nil", activity, err)
	}
}