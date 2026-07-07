package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"path"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lancer/log/internal/db"
	"github.com/lancer/log/internal/markdown"
	"github.com/lancer/log/internal/model"
	"github.com/lancer/log/internal/repo"
)

type PublicHandler struct {
	DB       *db.DB
	Tmpl     *Templates
	Settings []string
}

func NewPublicHandler(d *db.DB, t *Templates) *PublicHandler {
	return &PublicHandler{
		DB:       d,
		Tmpl:     t,
		Settings: []string{"branding", "nav", "footer", "hero", "stack", "about", "now", "archive", "shelf"},
	}
}

func (h *PublicHandler) siteData(ctx *gin.Context) (SiteData, map[string]json.RawMessage) {
	m, _ := repo.GetSettingsMap(ctx.Request.Context(), h.DB.Pool, h.Settings...)
	var s SiteData
	_ = DecodeSetting(m["branding"], &s.Brand)
	_ = DecodeSetting(m["nav"], &s.Nav)
	_ = DecodeSetting(m["footer"], &s.Footer)
	if s.Brand.Brand == "" {
		s.Brand.Brand = "lancer.log"
	}
	if s.Brand.FooterTag == "" {
		s.Brand.FooterTag = "Frontend practice, agent notes, and backend experiments."
	}
	if s.Brand.BuildBadge == "" {
		s.Brand.BuildBadge = "BUILT QUIETLY / NO TRACKING / NO ADS"
	}
	if s.Brand.CommitHash == "" {
		s.Brand.CommitHash = "a1f3c2d"
	}
	if s.Brand.SinceYear == 0 {
		s.Brand.SinceYear = 2026
	}
	if len(s.Nav.Links) == 0 {
		s.Nav = defaultNav()
	}
	if len(s.Footer.Cols) == 0 {
		s.Footer = defaultFooter()
	}
	s = removeSubscribeSurface(s)
	return s, m
}

func removeSubscribeSurface(s SiteData) SiteData {
	if strings.EqualFold(s.Nav.CTALabel, "subscribe") || s.Nav.CTAHref == "/about#contact" {
		s.Nav.CTALabel = ""
		s.Nav.CTAHref = ""
	}
	cols := make([]FooterCol, 0, len(s.Footer.Cols))
	for _, col := range s.Footer.Cols {
		if strings.EqualFold(col.H, "subscribe") {
			continue
		}
		links := make([]NavLink, 0, len(col.Links))
		for _, link := range col.Links {
			if link.Href == "/about#contact" || strings.Contains(strings.ToLower(link.Label), "subscribe") {
				continue
			}
			links = append(links, link)
		}
		col.Links = links
		cols = append(cols, col)
	}
	s.Footer.Cols = cols
	return s
}
func defaultNav() NavData {
	return NavData{
		Links: []NavLink{{"posts", "/"}, {"about", "/about"}, {"archive", "/archive"}, {"shelf", "/shelf"}},
	}
}

func defaultFooter() FooterData {
	return FooterData{Cols: []FooterCol{
		{H: "browse", Links: []NavLink{{"posts", "/"}, {"archive", "/archive"}, {"tags", "/tags"}, {"shelf", "/shelf"}}},
		{H: "about", Links: []NavLink{{"whoami", "/about"}}},
	}}
}

func (h *PublicHandler) Index(c *gin.Context) {
	site, m := h.siteData(c)
	var hero HeroData
	_ = DecodeSetting(m["hero"], &hero)
	var stack StackData
	_ = DecodeSetting(m["stack"], &stack)

	posts, _ := repo.ListPublished(c.Request.Context(), h.DB.Pool, 12)
	pinned, _ := repo.PinnedPost(c.Request.Context(), h.DB.Pool)
	if pinned == nil && len(posts) > 0 {
		p := posts[0]
		pinned = &p
	}

	data := IndexData{Site: site, Page: "posts", Hero: hero, Stack: stack, Pinned: pinned, Posts: posts, PostCount: len(posts)}
	h.render(c, "index", data)
}

func (h *PublicHandler) Post(c *gin.Context) {
	slug := c.Param("slug")
	p, err := repo.GetPostBySlug(c.Request.Context(), h.DB.Pool, slug)
	if errors.Is(err, repo.ErrNotFound) || (err == nil && !p.Published()) {
		h.NotFound(c)
		return
	}
	if err != nil {
		c.String(500, "err: %v", err)
		return
	}
	body := p.BodyHTML
	if body == "" && p.BodyMD != "" {
		body = markdown.Render(p.BodyMD)
	}

	site, _ := h.siteData(c)
	data := PostData{Site: site, Page: "", Post: p, AuthorName: "Lan"}
	data.Post.BodyHTML = body
	h.render(c, "post", data)
}

func (h *PublicHandler) About(c *gin.Context) {
	site, m := h.siteData(c)
	var about AboutData
	_ = DecodeSetting(m["about"], &about)
	if about.Title == "" || looksLikeLegacyAbout(about) {
		about = defaultAbout(site)
	}
	if about.BodyMarkdown != "" {
		about.BodyHTML = template.HTML(markdown.Render(about.BodyMarkdown))
	}
	var now NowData
	_ = DecodeSetting(m["now"], &now)
	if len(now.Lines) == 0 || looksLikeLegacyNow(now) {
		now = defaultNow()
	}
	data := AboutPageData{Site: site, Page: "about", About: about, Now: now}
	h.render(c, "about", data)
}

func (h *PublicHandler) Archive(c *gin.Context) {
	posts, err := repo.ListPublishedArchive(c.Request.Context(), h.DB.Pool)
	if err != nil {
		c.String(500, "archive: %v", err)
		return
	}
	h.renderArchive(c, posts, "", "", "archive", "No published posts yet.")
}

func (h *PublicHandler) Tags(c *gin.Context) {
	posts, err := repo.ListPublishedArchive(c.Request.Context(), h.DB.Pool)
	if err != nil {
		c.String(500, "tags: %v", err)
		return
	}
	h.renderArchive(c, posts, "tags", "all tags", "archive", "No tags yet.")
}

func (h *PublicHandler) Tag(c *gin.Context) {
	slug := c.Param("tag")
	posts, err := repo.ListPublishedByTag(c.Request.Context(), h.DB.Pool, slug)
	if err != nil {
		c.String(500, "tag: %v", err)
		return
	}
	label := slug
	if len(posts) > 0 {
		for _, tag := range posts[0].Tags {
			if tag.Slug == slug {
				label = tag.Name
				break
			}
		}
	}
	h.renderArchive(c, posts, "tag", "#"+label, "archive", "No published posts for this tag yet.")
}

func (h *PublicHandler) Section(c *gin.Context) {
	section := c.Param("section")
	posts, err := repo.ListPublishedBySection(c.Request.Context(), h.DB.Pool, section)
	if err != nil {
		c.String(500, "section: %v", err)
		return
	}
	h.renderArchive(c, posts, "section", section, "archive", "No published posts for this section yet.")
}

func (h *PublicHandler) Shelf(c *gin.Context) {
	site, m := h.siteData(c)
	var shelf ShelfData
	_ = DecodeSetting(m["shelf"], &shelf)
	if shelf.Title == "" {
		shelf = defaultShelf()
	}
	h.render(c, "shelf", ShelfPageData{Site: site, Page: "shelf", Shelf: shelf})
}

func (h *PublicHandler) renderArchive(c *gin.Context, posts []model.Post, filterKind, filterLabel, page, empty string) {
	site, m := h.siteData(c)
	var intro PageIntroData
	_ = DecodeSetting(m["archive"], &intro)
	if intro.Title == "" {
		intro = defaultArchiveIntro()
	}
	if filterLabel != "" {
		intro.EyebrowCmd = fmt.Sprintf("grep %q posts.json", filterLabel)
		intro.Title = filterLabel
		intro.TitleAccent = ""
		intro.Intro = fmt.Sprintf("Filtered view for %s. Use the side index to jump back into the full archive.", filterLabel)
	}
	tags, _ := repo.TagCounts(c.Request.Context(), h.DB.Pool)
	sections, _ := repo.SectionCounts(c.Request.Context(), h.DB.Pool)
	data := ArchivePageData{
		Site: site, Page: page, Archive: intro, Years: groupPostsByYear(posts), Posts: posts,
		Tags: toCountItems(tags), Sections: toCountItems(sections), TotalPosts: len(posts),
		TotalWords: totalWords(posts), FilterKind: filterKind, FilterLabel: filterLabel, EmptyText: empty,
	}
	h.render(c, "archive", data)
}

func groupPostsByYear(posts []model.Post) []ArchiveYear {
	index := map[string]int{}
	var years []ArchiveYear
	for _, post := range posts {
		year := "draft"
		if post.PublishedAt != nil {
			year = post.PublishedAt.Format("2006")
		}
		pos, ok := index[year]
		if !ok {
			index[year] = len(years)
			years = append(years, ArchiveYear{Year: year})
			pos = len(years) - 1
		}
		years[pos].Posts = append(years[pos].Posts, post)
	}
	sort.SliceStable(years, func(i, j int) bool { return years[i].Year > years[j].Year })
	return years
}

func totalWords(posts []model.Post) int {
	total := 0
	for _, post := range posts {
		total += post.Words
	}
	return total
}

func toCountItems[T interface {
	model.TagCount | model.SectionCount
}](items []T) []CountItem {
	out := make([]CountItem, 0, len(items))
	for _, item := range items {
		b, _ := json.Marshal(item)
		var count CountItem
		_ = json.Unmarshal(b, &count)
		out = append(out, count)
	}
	return out
}

func looksLikeLegacyAbout(about AboutData) bool {
	return strings.Contains(about.Title, "\u6d93") || about.BioYml.Role == "backend / infra"
}

func looksLikeLegacyNow(now NowData) bool {
	return len(now.Lines) > 0 && !now.Lines[0].IsCmd && now.Lines[0].K == "" && now.Lines[0].V == ""
}

func defaultArchiveIntro() PageIntroData {
	return PageIntroData{
		EyebrowCmd:  "git log --reverse --stat",
		Title:       "Archive",
		TitleAccent: "Archive",
		Intro:       "All published posts, grouped by time. This is a long-running learning log for frontend work, agents, backend practice, and ideas still taking shape.",
		Meta:        []AboutMeta{{K: "mode", V: "timeline"}, {K: "order", V: "newest first"}, {K: "status", V: "published only"}},
	}
}

func defaultAbout(site SiteData) AboutData {
	return AboutData{
		Title:       "A CS student moving from frontend toward agents and backend.",
		TitleAccent: "agents",
		Intro: []string{
			"I am a computer science undergraduate about to enter senior year, currently working as a frontend intern.",
			"This blog records the process of turning docs, internships, small projects, and self-study into practical engineering judgment.",
		},
		Meta:         []AboutMeta{{K: "role", V: "frontend intern"}, {K: "major", V: "computer science"}, {K: "next", V: "agent / backend"}, {K: "blog", V: site.Brand.Brand}},
		BioYml:       BioYml{Name: "Lan", Role: "frontend intern / CS student", Stack: "react / go / agent", Writes: "learning in public", Based: "china", Hosting: "self-hosted blog"},
		Uptime:       "still learning",
		BodyMarkdown: "## About me\n\nI am in that useful stage where concepts are starting to make sense, but real implementation still exposes gaps. This blog keeps those gaps visible.\n\n## About this site\n\nThis is not a resume and not a tutorial archive. It is an engineering log for frontend internship notes, React practice, Go and backend experiments, agent ideas, and route changes in my learning plan.\n\n## Interests\n\nI like Max Verstappen's consistency and attack, and Stephen Curry's way of turning long practice into instinct. Coding has a bit of that too: repeat the basics, then stay calm in complex situations.",
	}
}

func defaultNow() NowData {
	return NowData{Lines: []NowLine{
		{IsCmd: true, F: "cat", Args: "now.txt", C: "# current focus"},
		{Arrow: "->", K: "internship", V: "frontend", IsString: true},
		{Arrow: "->", K: "learning", V: "react / go / agents", IsString: true},
		{Arrow: "->", K: "next", V: "ship small projects", IsString: true},
	}}
}

func defaultShelf() ShelfData {
	return ShelfData{
		EyebrowCmd:  "ls shelf/learning-stack",
		Title:       "Shelf",
		TitleAccent: "Shelf",
		Intro:       "A living shelf for books, docs, tools, and courses I am reading, using, or planning to study more deeply.",
		Meta:        []AboutMeta{{K: "type", V: "books / tools / courses"}, {K: "update", V: "manual"}, {K: "bias", V: "small useful things"}},
		Groups: []ShelfGroup{
			{Title: "Learning", Eyebrow: "reading queue", Desc: "Resources that strengthen the engineering base.", Items: []ShelfItem{
				{Title: "React Docs", Desc: "Rebuild muscle memory around components, state, effects, and data flow.", Meta: "frontend", Href: "https://react.dev/", Status: "re-reading", Tags: []string{"react", "frontend"}},
				{Title: "Go by Example", Desc: "Small examples for Go syntax and standard library practice.", Meta: "backend", Href: "https://gobyexample.com/", Status: "active", Tags: []string{"go", "backend"}},
			}},
			{Title: "Tools", Eyebrow: "daily kit", Desc: "Tools used in internship work and personal projects.", Items: []ShelfItem{
				{Title: "VS Code / Cursor", Desc: "Main entry for frontend work, reading Go projects, and agent collaboration.", Meta: "editor", Href: "#", Status: "daily", Tags: []string{"editor", "agent"}},
				{Title: "Postman / Apifox", Desc: "API debugging and frontend/backend integration practice.", Meta: "api", Href: "#", Status: "daily", Tags: []string{"api", "backend"}},
			}},
		},
	}
}

func (h *PublicHandler) NotFound(c *gin.Context) {
	site, _ := h.siteData(c)
	data := NotFoundData{Site: site, Page: "", Path: c.Request.URL.Path}
	c.Status(404)
	if err := h.Tmpl.Execute(c.Writer, "notfound", data); err != nil {
		c.String(500, "404 template error: %v", err)
	}
}

func (h *PublicHandler) render(c *gin.Context, name string, data any) {
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Tmpl.Execute(c.Writer, name, data); err != nil {
		c.String(500, "template error: %v", err)
	}
}

func (h *PublicHandler) Static(r *gin.Engine) {
	r.Static("/static", path.Join("web", "static"))
}

func ensureNoAdminLeak(s string) bool { return strings.HasPrefix(s, "/admin") }
