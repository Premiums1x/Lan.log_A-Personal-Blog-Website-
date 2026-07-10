package server

import (
	"context"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lancer/log/internal/auth"
	"github.com/lancer/log/internal/config"
	"github.com/lancer/log/internal/db"
	"github.com/lancer/log/internal/github"
	"github.com/lancer/log/internal/handler"
	"github.com/lancer/log/web"
)

func loadTemplates() (*handler.Templates, error) {
	tplFS, err := fs.Sub(web.TemplatesFS, "templates")
	if err != nil {
		return nil, err
	}
	funcs := handler.FuncMap()
	for k, v := range handler.MarkdownFuncs() {
		funcs[k] = v
	}
	base := []string{"layout.tmpl"}
	pages := map[string]string{
		"index":    "index.tmpl",
		"post":     "post.tmpl",
		"about":    "about.tmpl",
		"archive":  "archive.tmpl",
		"shelf":    "shelf.tmpl",
		"notfound": "notfound.tmpl",
	}
	return handler.LoadTemplates(tplFS, base, pages, funcs)
}

type Server struct {
	Cfg config.Config
	DB  *db.DB
	Gin *gin.Engine
}

func New(ctx context.Context, cfg config.Config) (*Server, error) {
	d, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	tmpl, err := loadTemplates()
	if err != nil {
		return nil, err
	}
	mgr := auth.NewManager(cfg.JWTSecret, cfg.JWTTTLHours)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	staticFS, _ := fs.Sub(web.StaticFS, "static")
	r.StaticFS("/static", http.FS(staticFS))

	var gh *github.Client
	if cfg.GitHub.Token != "" {
		gh = github.NewClient(cfg.GitHub.Token, cfg.GitHub.Username)
	}

	pub := handler.NewPublicHandler(d, tmpl, gh)
	api := handler.NewAPIHandler(d, mgr, cfg)

	r.GET("/", pub.Index)
	r.GET("/posts/:slug", pub.Post)
	r.GET("/about", pub.About)
	r.GET("/archive", pub.Archive)
	r.GET("/tags", pub.Tags)
	r.GET("/tags/:tag", pub.Tag)
	r.GET("/section/:section", pub.Section)
	r.GET("/shelf", pub.Shelf)

	ag := r.Group("/api")
	ag.POST("/login", api.Login)
	ag.GET("/brand", api.Brand)
	ag.POST("/password-reset/request", api.RequestPasswordReset)
	ag.POST("/password-reset/confirm", api.ConfirmPasswordReset)
	authed := ag.Group("")
	authed.Use(mgr.Middleware())
	authed.GET("/me", api.Me)
	authed.GET("/account", api.Account)
	authed.PUT("/account/recovery-email", api.UpdateRecoveryEmail)
	authed.PUT("/account/password", api.UpdateAccountPassword)
	authed.GET("/posts", api.ListPosts)
	authed.GET("/posts/:id", api.GetPost)
	authed.POST("/posts", api.CreatePost)
	authed.PUT("/posts/:id", api.UpdatePost)
	authed.DELETE("/posts/:id", api.DeletePost)
	authed.GET("/settings", api.ListSettings)
	authed.GET("/settings/:key", api.GetSetting)
	authed.PUT("/settings/:key", api.SetSetting)

	adminFS := openAdminFS()
	if adminFS != nil {
		r.NoRoute(func(c *gin.Context) {
			p := c.Request.URL.Path
			if strings.HasPrefix(p, "/admin") {
				serveAdmin(adminFS, c)
				return
			}
			pub.NotFound(c)
		})
	} else {
		r.NoRoute(pub.NotFound)
	}

	return &Server{Cfg: cfg, DB: d, Gin: r}, nil
}

func (s *Server) Run() error { return s.Gin.Run(s.Cfg.HTTPAddr) }
func (s *Server) Close()     { s.DB.Close() }

func serveAdmin(fsys fs.FS, c *gin.Context) {
	rest := strings.TrimPrefix(c.Request.URL.Path, "/admin")
	rest = strings.TrimPrefix(rest, "/")
	if rest == "" || rest == "/" {
		rest = "index.html"
	}
	if _, err := fs.Stat(fsys, rest); err != nil {
		rest = "index.html"
	}
	http.ServeFileFS(c.Writer, c.Request, fsys, rest)
}

func openAdminFS() fs.FS {
	return adminFS()
}
