package handler

import (
	"encoding/json"
	"html/template"
	"time"

	"github.com/lancer/log/internal/model"
)

// ---------- Settings DTOs decoded from the jsonb settings table ----------

type BrandData struct {
	Brand      string `json:"brand"`
	FooterTag  string `json:"footer_tag"`
	SinceYear  int    `json:"since_year"`
	CommitHash string `json:"commit_hash"`
	BuildBadge string `json:"build_badge"`
}

type NavLink struct {
	Label string `json:"label"`
	Href  string `json:"href"`
}

type NavData struct {
	Links    []NavLink `json:"links"`
	CTALabel string    `json:"cta_label"`
	CTAHref  string    `json:"cta_href"`
}

type CornerStat struct{ Label, Val string }
type HeroMeta struct{ K, V string }
type HeroData struct {
	EyebrowCmd  string       `json:"eyebrow_cmd"`
	Title       string       `json:"title"`
	TitleAccent string       `json:"title_accent"`
	TitleTail   string       `json:"title_tail"`
	Sub         string       `json:"sub"`
	Meta        []HeroMeta   `json:"meta"`
	Corner      []CornerStat `json:"corner"`
}

type StackCell struct{ Ic, Title, Desc string }
type StackData struct {
	Cells []StackCell `json:"cells"`
}

type FooterCol struct {
	H     string    `json:"h"`
	Links []NavLink `json:"links"`
}
type FooterData struct {
	Cols []FooterCol `json:"cols"`
}

type AboutMeta struct{ K, V string }
type BioYml struct {
	Name, Role, Stack, Writes, Based, Hosting string
}
type AboutData struct {
	Title        string        `json:"title"`
	TitleAccent  string        `json:"title_accent"`
	Intro        []string      `json:"intro"`
	Meta         []AboutMeta   `json:"meta"`
	BioYml       BioYml        `json:"bio_yml"`
	Uptime       string        `json:"uptime"`
	BodyMarkdown string        `json:"body_md"`
	BodyHTML     template.HTML `json:"-"`
}

type NowLine struct {
	IsCmd    bool   `json:"is_cmd"`
	F        string `json:"f"`
	Args     string `json:"args"`
	C        string `json:"c"`
	Arrow    string `json:"arrow"`
	K        string `json:"k"`
	V        string `json:"v"`
	IsString bool   `json:"is_string"`
}
type NowData struct {
	Lines []NowLine `json:"lines"`
}
type PageIntroData struct {
	EyebrowCmd  string      `json:"eyebrow_cmd"`
	Title       string      `json:"title"`
	TitleAccent string      `json:"title_accent"`
	Intro       string      `json:"intro"`
	Meta        []AboutMeta `json:"meta"`
}

type CountItem struct {
	Slug  string
	Name  string
	Count int
}

type ArchiveYear struct {
	Year  string
	Posts []model.Post
}

type ShelfItem struct {
	Title  string   `json:"title"`
	Desc   string   `json:"desc"`
	Meta   string   `json:"meta"`
	Href   string   `json:"href"`
	Status string   `json:"status"`
	Tags   []string `json:"tags"`
}

type ShelfGroup struct {
	Title   string      `json:"title"`
	Eyebrow string      `json:"eyebrow"`
	Desc    string      `json:"desc"`
	Items   []ShelfItem `json:"items"`
}

type ShelfData struct {
	EyebrowCmd  string       `json:"eyebrow_cmd"`
	Title       string       `json:"title"`
	TitleAccent string       `json:"title_accent"`
	Intro       string       `json:"intro"`
	Meta        []AboutMeta  `json:"meta"`
	Groups      []ShelfGroup `json:"groups"`
}

// ---------- Page payloads ----------

type SiteData struct {
	Brand  BrandData
	Nav    NavData
	Footer FooterData
}

type IndexData struct {
	Site      SiteData
	Page      string
	Hero      HeroData
	Stack     StackData
	Pinned    *model.Post
	Posts     []model.Post
	PostCount int
}

type HeatmapCell struct {
	Date  string
	Count int
	Level int
}

type HeatmapData struct {
	Username string
	URL      string
	Total    int
	Weeks    [][]HeatmapCell
}

type PostData struct {
	Site       SiteData
	Page       string
	Post       model.Post
	Prev, Next *model.Post
	AuthorName string
	Heatmap    *HeatmapData
}

type AboutPageData struct {
	Site  SiteData
	Page  string
	About AboutData
	Now   NowData
}

type ArchivePageData struct {
	Site        SiteData
	Page        string
	Archive     PageIntroData
	Years       []ArchiveYear
	Posts       []model.Post
	Tags        []CountItem
	Sections    []CountItem
	TotalPosts  int
	TotalWords  int
	FilterKind  string
	FilterLabel string
	EmptyText   string
}

type ShelfPageData struct {
	Site  SiteData
	Page  string
	Shelf ShelfData
}

type NotFoundData struct {
	Site SiteData
	Page string
	Path string
}

// DecodeSetting unmarshals a json.RawMessage into a typed settings struct.
func DecodeSetting(raw json.RawMessage, dest any) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, &dest)
}

// ---------- Template helpers ----------

func FuncMap() template.FuncMap {
	return template.FuncMap{
		"FormatDate":   formatDate,
		"YearNow":      func() int { return time.Now().Year() },
		"CommitShort":  commitShort,
		"FirstChar":    firstChar,
		"RenderHTML":   func(s string) template.HTML { return template.HTML(s) },
		"RenderInline": renderInline,
		"SplitAccent":  splitAccent,
	}
}

func formatDate(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return t.Format("2006-01-02")
}
