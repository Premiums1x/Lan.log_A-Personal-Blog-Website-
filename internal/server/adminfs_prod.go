//go:build admin

package server

import (
	"io/fs"
	"strings"

	"github.com/lancer/log/web"
)

// prod: admin SPA embedded at build time (run `npm run build` in web/admin first).
func adminFS() fs.FS {
	sub, err := fs.Sub(web.AdminFS, "admin-dist")
	if err != nil {
		return nil
	}
	return strippedFS{FS: sub}
}

// strippedFS rewrites paths so that "index.html" and "assets/*" are addressable
// from /admin/* without the leading "admin/". We already Sub to admin-dist root.
type strippedFS struct{ fs.FS }

func (s strippedFS) Open(name string) (fs.File, error) {
	name = strings.TrimPrefix(name, "/")
	if name == "" {
		name = "index.html"
	}
	return s.FS.Open(name)
}