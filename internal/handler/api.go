package handler

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lancer/log/internal/auth"
	"github.com/lancer/log/internal/config"
	"github.com/lancer/log/internal/db"
	"github.com/lancer/log/internal/markdown"
	"github.com/lancer/log/internal/repo"
)

type APIHandler struct {
	DB     *db.DB
	Auth   *auth.Manager
	Config config.Config
}

func NewAPIHandler(d *db.DB, m *auth.Manager, cfg config.Config) *APIHandler {
	return &APIHandler{DB: d, Auth: m, Config: cfg}
}

func ok(c *gin.Context, v any)                  { c.JSON(200, v) }
func fail(c *gin.Context, code int, msg string) { c.JSON(code, gin.H{"error": msg}) }

// Brand returns the public brand name (from settings/branding.brand).
// Public endpoint - no JWT required - used by admin SPA shell/login.
func (h *APIHandler) Brand(c *gin.Context) {
	var raw []byte
	err := h.DB.Pool.QueryRow(c.Request.Context(),
		`SELECT value::text FROM settings WHERE section_key='branding'`).Scan(&raw)
	if err != nil {
		ok(c, gin.H{"brand": ""})
		return
	}
	var v struct {
		Brand string `json:"brand"`
	}
	if json.Unmarshal(raw, &v) != nil {
		ok(c, gin.H{"brand": ""})
		return
	}
	ok(c, gin.H{"brand": v.Brand})
}

// ---------- Auth ----------

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *APIHandler) Login(c *gin.Context) {
	var r loginReq
	if err := c.ShouldBindJSON(&r); err != nil {
		fail(c, 400, "bad request")
		return
	}
	u, err := repo.GetUserByUsername(c.Request.Context(), h.DB.Pool, strings.TrimSpace(r.Username))
	if errors.Is(err, repo.ErrNotFound) {
		fail(c, 401, "用户名或密码错误")
		return
	}
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	if !auth.CheckPassword(u.PasswordHash, r.Password) {
		fail(c, 401, "用户名或密码错误")
		return
	}
	token, err := h.Auth.Issue(u)
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	ok(c, gin.H{"token": token, "user": gin.H{"id": u.ID, "username": u.Username, "display_name": u.DisplayName}})
}

func (h *APIHandler) Me(c *gin.Context) {
	uid := c.MustGet("uid").(uuid.UUID)
	un := c.MustGet("username").(string)
	ok(c, gin.H{"id": uid, "username": un})
}

// ---------- Posts ----------

func (h *APIHandler) ListPosts(c *gin.Context) {
	posts, err := repo.ListAll(c.Request.Context(), h.DB.Pool)
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	ok(c, gin.H{"items": posts, "total": len(posts)})
}

func (h *APIHandler) GetPost(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, 400, "bad id")
		return
	}
	p, err := repo.GetPostByID(c.Request.Context(), h.DB.Pool, id)
	if errors.Is(err, repo.ErrNotFound) {
		fail(c, 404, "not found")
		return
	}
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	ok(c, p)
}

type postReq struct {
	Slug     string   `json:"slug"`
	Title    string   `json:"title"`
	Excerpt  string   `json:"excerpt"`
	BodyMD   string   `json:"body_md"`
	CoverURL string   `json:"cover_url"`
	Section  string   `json:"section"`
	Status   string   `json:"status"`
	Pinned   bool     `json:"pinned"`
	TagNames []string `json:"tag_names"`
}

func (h *APIHandler) CreatePost(c *gin.Context) {
	var r postReq
	if err := c.ShouldBindJSON(&r); err != nil {
		fail(c, 400, err.Error())
		return
	}
	if err := validatePost(r); err != nil {
		fail(c, 400, err.Error())
		return
	}
	bodyHTML := markdown.Render(r.BodyMD)
	words := markdown.WordCount(r.BodyMD)
	readMin := markdown.ReadMinutes(r.BodyMD)
	excerpt := strings.TrimSpace(r.Excerpt)
	if excerpt == "" {
		excerpt = markdown.Excerpt(r.BodyMD, 120)
	}
	commit := newCommitHash(r.Slug)
	section := strings.TrimSpace(r.Section)
	if section == "" {
		section = "posts"
	}
	p, err := repo.CreatePost(c.Request.Context(), h.DB.Pool, repo.PostInput{
		Slug: r.Slug, Title: r.Title, Excerpt: excerpt, BodyMD: r.BodyMD, BodyHTML: bodyHTML,
		CoverURL: r.CoverURL, Section: section, Status: r.Status, Pinned: r.Pinned, TagNames: r.TagNames,
	}, commit, readMin, words)
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	ok(c, p)
}

func (h *APIHandler) UpdatePost(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, 400, "bad id")
		return
	}
	var r postReq
	if err := c.ShouldBindJSON(&r); err != nil {
		fail(c, 400, err.Error())
		return
	}
	if err := validatePost(r); err != nil {
		fail(c, 400, err.Error())
		return
	}
	bodyHTML := markdown.Render(r.BodyMD)
	words := markdown.WordCount(r.BodyMD)
	readMin := markdown.ReadMinutes(r.BodyMD)
	excerpt := strings.TrimSpace(r.Excerpt)
	if excerpt == "" {
		excerpt = markdown.Excerpt(r.BodyMD, 120)
	}
	commit := newCommitHash(r.Slug)
	section := strings.TrimSpace(r.Section)
	if section == "" {
		section = "posts"
	}
	p, err := repo.UpdatePost(c.Request.Context(), h.DB.Pool, id, repo.PostInput{
		Slug: r.Slug, Title: r.Title, Excerpt: excerpt, BodyMD: r.BodyMD, BodyHTML: bodyHTML,
		CoverURL: r.CoverURL, Section: section, Status: r.Status, Pinned: r.Pinned, TagNames: r.TagNames,
	}, commit, readMin, words)
	if errors.Is(err, repo.ErrNotFound) {
		fail(c, 404, "not found")
		return
	}
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	ok(c, p)
}

func (h *APIHandler) DeletePost(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, 400, "bad id")
		return
	}
	if err := repo.DeletePost(c.Request.Context(), h.DB.Pool, id); err != nil {
		fail(c, 400, err.Error())
		return
	}
	ok(c, gin.H{"deleted": true})
}

func validatePost(r postReq) error {
	r.Slug = strings.TrimSpace(r.Slug)
	r.Title = strings.TrimSpace(r.Title)
	if r.Slug == "" {
		return errors.New("slug is required")
	}
	if r.Title == "" {
		return errors.New("title is required")
	}
	if r.Status != "draft" && r.Status != "published" {
		return errors.New("status must be draft or published")
	}
	return nil
}

// ---------- Settings ----------

func (h *APIHandler) ListSettings(c *gin.Context) {
	keys, err := repo.ListSettingKeys(c.Request.Context(), h.DB.Pool)
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	m, _ := repo.GetSettingsMap(c.Request.Context(), h.DB.Pool, keys...)
	out := make(map[string]any, len(m))
	for k, v := range m {
		var val any
		_ = json.Unmarshal(v, &val)
		out[k] = val
	}
	ok(c, gin.H{"keys": keys, "settings": out})
}

func (h *APIHandler) GetSetting(c *gin.Context) {
	key := c.Param("key")
	var raw []byte
	err := h.DB.Pool.QueryRow(c.Request.Context(), `SELECT value::text FROM settings WHERE section_key=$1`, key).Scan(&raw)
	if err != nil {
		ok(c, gin.H{"section_key": key, "value": gin.H{}, "updated_at": time.Now()})
		return
	}
	var val any
	_ = json.Unmarshal(raw, &val)
	ok(c, gin.H{"section_key": key, "value": val, "updated_at": time.Now()})
}

type settingReq struct {
	Value any `json:"value"`
}

func (h *APIHandler) SetSetting(c *gin.Context) {
	key := c.Param("key")
	var r settingReq
	if err := c.ShouldBindJSON(&r); err != nil {
		fail(c, 400, err.Error())
		return
	}
	if err := repo.SetSetting(c.Request.Context(), h.DB.Pool, key, r.Value); err != nil {
		fail(c, 400, err.Error())
		return
	}
	ok(c, gin.H{"section_key": key, "saved": true})
}
