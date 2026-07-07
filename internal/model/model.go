package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	DisplayName  string    `json:"display_name"`
	CreatedAt    time.Time `json:"created_at"`
}

type Post struct {
	ID          uuid.UUID  `json:"id"`
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Excerpt     string     `json:"excerpt"`
	BodyMD      string     `json:"body_md"`
	BodyHTML    string     `json:"body_html"`
	CoverURL    string     `json:"cover_url"`
	Section     string     `json:"section"`
	Status      string     `json:"status"`
	CommitHash  string     `json:"commit_hash"`
	ReadMinutes int        `json:"read_minutes"`
	Words       int        `json:"words"`
	Pinned      bool       `json:"pinned"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Tags        []Tag      `json:"tags"`
}

func (p Post) Published() bool { return p.Status == "published" }

type Tag struct {
	ID   uuid.UUID `json:"id"`
	Slug string    `json:"slug"`
	Name string    `json:"name"`
}

type TagCount struct {
	Slug  string `json:"slug"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type SectionCount struct {
	Slug  string `json:"slug"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type Setting struct {
	SectionKey string    `json:"section_key"`
	Value      any       `json:"value"`
	UpdatedAt  time.Time `json:"updated_at"`
}
