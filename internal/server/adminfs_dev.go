//go:build !admin

package server

import "io/fs"

// dev mode: admin SPA served by Vite (npm run dev) on :5174 which proxies /api here.
func adminFS() fs.FS { return nil }