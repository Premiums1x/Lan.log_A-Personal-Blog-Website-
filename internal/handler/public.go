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
	"github.com/lancer/log/internal/github"
	"github.com/lancer/log/internal/markdown"
	"github.com/lancer/log/internal/model"
	"github.com/lancer/log/internal/repo"
)

type PublicHandler struct {
	DB       *db.DB
	Tmpl     *Templates
	Settings []string
	GitHub   *github.Client
}

func NewPublicHandler(d *db.DB, t *Templates, gh *github.Client) *PublicHandler {
	return &PublicHandler{
		DB:       d,
		Tmpl:     t,
		Settings: []string{"branding", "nav", "footer", "hero", "stack", "about", "now", "archive", "shelf"},
		GitHub:   gh,
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
	data := PostData{Site: site, Page: "", Post: p, AuthorName: "Lancer"}
	data.Post.BodyHTML = body
	if h.GitHub != nil && h.GitHub.Enabled() {
		if cal, err := h.GitHub.FetchContributions(c.Request.Context()); err == nil {
			data.Heatmap = calToHeatmap(cal)
		} else {
			fmt.Printf("github heatmap error: %v\n", err)
		}
	} else {
		fmt.Println("github heatmap: token not configured or disabled")
	}
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
		EyebrowCmd:  "season --all --newest-first",
		Title:       "Season Log",
		TitleAccent: "Season",
		Intro:       "不是胜场统计，而是每一阶段真实留下的路线。回头看时，希望能看见自己怎样探索、失误、修正，然后继续向前。",
		Meta:        []AboutMeta{{K: "mode", V: "season record"}, {K: "order", V: "newest first"}, {K: "status", V: "published rounds"}},
	}
}
func defaultAbout(site SiteData) AboutData {
	return AboutData{
		Title:       "有些路还没走明白，但我仍然相信更好的明天。",
		TitleAccent: "更好的明天",
		Intro: []string{
			"我是 Lancer。这里不公开现实身份，只记录我愿意留下的思考、热爱和正在走的路线。",
			"我喜欢进攻带来的突破，也希望自己在真正的残局里足够冷静，让看到这一回合的人感到放心。",
		},
		Meta:         []AboutMeta{{K: "callsign", V: "Lancer"}, {K: "mode", V: "exploring"}, {K: "writes", V: "field notes"}, {K: "blog", V: site.Brand.Brand}},
		BioYml:       BioYml{Name: "Lancer", Role: "explorer / builder", Stack: "frontend / go / agent", Writes: "field notes", Based: "private", Hosting: "self-hosted"},
		Uptime:       "这一回合没结束",
		BodyMarkdown: "## 为什么是这个博客\n\n它首先写给我自己。不是为了把一段经历包装成漂亮结论，而是为了保留探索、失误、修正和继续向前的证据。\n\n## 进攻与残局\n\n我喜欢突破手撕开防线的瞬间，也喜欢残局里噪音逐渐消失、只剩判断和执行的时刻。写代码也很像这样：有时要大胆进入，有时要让自己慢下来，把问题一个个处理。\n\n## 更好的明天\n\n生活并不总按计划推进。偶尔不顺利时，我希望自己还能咬咬牙，把今天向前推一点。只要这一回合没有结束，就还有下一次选择。",
	}
}
func defaultNow() NowData {
	return NowData{Lines: []NowLine{
		{IsCmd: true, F: "cat", Args: "round.txt", C: "# current round"},
		{Arrow: "->", K: "building", V: "personal blog / field notes", IsString: true},
		{Arrow: "->", K: "learning", V: "frontend / go / agents", IsString: true},
		{Arrow: "->", K: "mental", V: "这一回合没结束", IsString: true},
		{Arrow: "->", K: "next", V: "把今天推进一点", IsString: true},
	}}
}
func defaultShelf() ShelfData {
	return ShelfData{
		EyebrowCmd:  "loadout --show active",
		Title:       "Loadout",
		TitleAccent: "Loadout",
		Intro:       "带进下一回合的工具、文档、故事、比赛和声音。它们不都直接关于代码，但都在塑造我的判断和热情。",
		Meta:        []AboutMeta{{K: "type", V: "tools / stories / signals"}, {K: "update", V: "manual"}, {K: "owner", V: "Lancer"}},
		Groups: []ShelfGroup{
			{Title: "Engineering", Eyebrow: "active tools", Desc: "帮助我把想法变成真实结果的工具和资料。", Items: []ShelfItem{
				{Title: "React Docs", Desc: "回到组件、状态、Effect 与数据流的原点。", Meta: "frontend", Href: "https://react.dev/", Status: "active", Tags: []string{"react", "frontend"}},
				{Title: "Go by Example", Desc: "用小例子训练后端语法、标准库和工程直觉。", Meta: "backend", Href: "https://gobyexample.com/", Status: "active", Tags: []string{"go", "backend"}},
			}},
			{Title: "Competition", Eyebrow: "mental models", Desc: "从赛道、球场和服务器里学到的进攻与冷静。", Items: []ShelfItem{
				{Title: "Racecraft", Desc: "路线、节奏、轮胎与风险选择。速度来自精确，也来自敢于行动。", Meta: "F1", Status: "watching", Tags: []string{"Verstappen", "Leclerc"}},
				{Title: "Clutch Notes", Desc: "突破先创造空间，残局再把噪音降到最低。", Meta: "Counter-Strike", Status: "training", Tags: []string{"NiKo", "clutch"}},
			}},
			{Title: "Soundtrack", Eyebrow: "energy source", Desc: "为长时间投入和想象中的高光时刻提供节奏。", Items: []ShelfItem{
				{Title: "Millennium Mandopop", Desc: "周杰伦、邓紫棋，以及那些能把普通夜晚变成一段故事的歌。", Meta: "playlist", Status: "repeat", Tags: []string{"Mandopop", "2000s"}},
				{Title: "Phonk / EDM", Desc: "快节奏、低频和推进感，适合需要重新进入状态的时刻。", Meta: "focus mix", Status: "playing", Tags: []string{"phonk", "edm"}},
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

func calToHeatmap(cal *github.Calendar) *HeatmapData {
	hd := &HeatmapData{
		Username: cal.Login,
		URL:      cal.URL,
		Total:    cal.Total,
	}
	for _, week := range cal.Weeks {
		cells := make([]HeatmapCell, 0, len(week))
		for _, d := range week {
			cells = append(cells, HeatmapCell{
				Date:  d.Date,
				Count: d.Count,
				Level: heatmapLevel(d.Count),
			})
		}
		hd.Weeks = append(hd.Weeks, cells)
	}
	return hd
}

func heatmapLevel(n int) int {
	switch {
	case n == 0:
		return 0
	case n <= 2:
		return 1
	case n <= 5:
		return 2
	case n <= 8:
		return 3
	default:
		return 4
	}
}
