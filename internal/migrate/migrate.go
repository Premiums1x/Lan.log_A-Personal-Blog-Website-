package migrate

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/lancer/log/web"
)

// Run applies all up-migration .sql files under web/migrations in lexical order.
// Idempotent: wraps each file in a transaction and skips on harmless re-runs
// for objects that already exist (our migrations use IF NOT EXISTS / ON CONFLICT).
func Run(ctx context.Context, dsn string) error {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.Close(ctx)

	files, err := fs.ReadDir(web.MigrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	var sqls []string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".sql") {
			sqls = append(sqls, f.Name())
		}
	}
	sort.Strings(sqls)

	for _, name := range sqls {
		b, err := web.MigrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		if _, err := conn.Exec(ctx, string(b)); err != nil {
			return fmt.Errorf("apply %s: %w", name, err)
		}
		fmt.Printf("  ✓ migration %s\n", name)
	}
	return nil
}