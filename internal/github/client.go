package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	token    string
	username string
	http     *http.Client

	cacheMu  sync.RWMutex
	cached   *Calendar
	cachedAt time.Time
	ttl      time.Duration
}

type Calendar struct {
	Login string
	URL   string
	Total int
	Weeks [][]Day
}

type Day struct {
	Date  string
	Count int
}

func NewClient(token, username string) *Client {
	return &Client{
		token:    token,
		username: username,
		http:     &http.Client{Timeout: 15 * time.Second},
		ttl:      1 * time.Hour,
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.token != ""
}

const contributionsQuery = `query {
  viewer {
    login
    url
    contributionsCollection {
      contributionCalendar {
        totalContributions
        weeks {
          contributionDays {
            contributionCount
            date
          }
        }
      }
    }
  }
}`

type graphQLResponse struct {
	Data struct {
		Viewer struct {
			Login                   string `json:"login"`
			URL                     string `json:"url"`
			ContributionsCollection struct {
				ContributionCalendar struct {
					TotalContributions int `json:"totalContributions"`
					Weeks              []struct {
						ContributionDays []struct {
							ContributionCount int    `json:"contributionCount"`
							Date              string `json:"date"`
						} `json:"contributionDays"`
					} `json:"weeks"`
				} `json:"contributionCalendar"`
			} `json:"contributionsCollection"`
		} `json:"viewer"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (c *Client) FetchContributions(ctx context.Context) (*Calendar, error) {
	c.cacheMu.RLock()
	if c.cached != nil && time.Since(c.cachedAt) < c.ttl {
		cal := c.cached
		c.cacheMu.RUnlock()
		return cal, nil
	}
	c.cacheMu.RUnlock()

	body, _ := json.Marshal(map[string]string{"query": contributionsQuery})
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.github.com/graphql", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("github graphql: HTTP %d: %s", resp.StatusCode, string(raw))
	}

	var gql graphQLResponse
	if err := json.Unmarshal(raw, &gql); err != nil {
		return nil, fmt.Errorf("parse github response: %w", err)
	}
	if len(gql.Errors) > 0 {
		return nil, fmt.Errorf("github graphql error: %s", gql.Errors[0].Message)
	}

	cal := &Calendar{
		Login: gql.Data.Viewer.Login,
		URL:   gql.Data.Viewer.URL,
		Total: gql.Data.Viewer.ContributionsCollection.ContributionCalendar.TotalContributions,
	}
	if c.username != "" {
		cal.Login = c.username
	}

	for _, w := range gql.Data.Viewer.ContributionsCollection.ContributionCalendar.Weeks {
		days := make([]Day, 0, len(w.ContributionDays))
		for _, d := range w.ContributionDays {
			days = append(days, Day{Date: d.Date, Count: d.ContributionCount})
		}
		cal.Weeks = append(cal.Weeks, days)
	}

	c.cacheMu.Lock()
	c.cached = cal
	c.cachedAt = time.Now()
	c.cacheMu.Unlock()

	return cal, nil
}