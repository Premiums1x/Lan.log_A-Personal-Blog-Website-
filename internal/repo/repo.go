package repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lancer/log/internal/db"
	"github.com/lancer/log/internal/model"
)

var ErrNotFound = errors.New("not found")

// ---------- Users ----------

const userCols = `id, username, password_hash, display_name, COALESCE(recovery_email,''), created_at, COALESCE(password_updated_at, created_at)`

func GetUserByUsername(ctx context.Context, q db.Conn, username string) (model.User, error) {
	var u model.User
	err := q.QueryRow(ctx,
		`SELECT `+userCols+` FROM users WHERE username=$1`,
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.DisplayName, &u.RecoveryEmail, &u.CreatedAt, &u.PasswordUpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrNotFound
	}
	return u, err
}

func GetUserByID(ctx context.Context, q db.Conn, id uuid.UUID) (model.User, error) {
	var u model.User
	err := q.QueryRow(ctx,
		`SELECT `+userCols+` FROM users WHERE id=$1`,
		id,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.DisplayName, &u.RecoveryEmail, &u.CreatedAt, &u.PasswordUpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrNotFound
	}
	return u, err
}

func CountUsers(ctx context.Context, q db.Conn) (int, error) {
	var n int
	err := q.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&n)
	return n, err
}

func CreateUser(ctx context.Context, q db.Conn, username, hash, displayName string) (model.User, error) {
	return CreateUserWithEmail(ctx, q, username, hash, displayName, "")
}

func CreateUserWithEmail(ctx context.Context, q db.Conn, username, hash, displayName, recoveryEmail string) (model.User, error) {
	var u model.User
	err := q.QueryRow(ctx,
		`INSERT INTO users (username, password_hash, display_name, recovery_email, password_updated_at) VALUES ($1,$2,$3,$4,now())
		 RETURNING `+userCols,
		username, hash, displayName, recoveryEmail,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.DisplayName, &u.RecoveryEmail, &u.CreatedAt, &u.PasswordUpdatedAt)
	return u, err
}

func UpdateUserRecoveryEmail(ctx context.Context, q db.Conn, id uuid.UUID, email string) error {
	ct, err := q.Exec(ctx, `UPDATE users SET recovery_email=$2 WHERE id=$1`, id, email)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func UpdateUserPassword(ctx context.Context, q db.Conn, id uuid.UUID, hash string) error {
	ct, err := q.Exec(ctx, `UPDATE users SET password_hash=$2, password_updated_at=now() WHERE id=$1`, id, hash)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type PasswordResetCode struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	CodeHash  string
	ExpiresAt time.Time
	Attempts  int
	CreatedAt time.Time
}

func CountRecentPasswordResetCodes(ctx context.Context, q db.Conn, userID uuid.UUID, since time.Time) (int, error) {
	var n int
	err := q.QueryRow(ctx,
		`SELECT count(*) FROM password_reset_codes WHERE user_id=$1 AND created_at >= $2`,
		userID, since,
	).Scan(&n)
	return n, err
}

func CreatePasswordResetCode(ctx context.Context, q db.Conn, userID uuid.UUID, codeHash string, expiresAt time.Time, requestedIP string) error {
	_, err := q.Exec(ctx,
		`INSERT INTO password_reset_codes (user_id, code_hash, expires_at, requested_ip) VALUES ($1,$2,$3,$4)`,
		userID, codeHash, expiresAt, requestedIP,
	)
	return err
}

func LatestActivePasswordResetCode(ctx context.Context, q db.Conn, userID uuid.UUID) (PasswordResetCode, error) {
	var r PasswordResetCode
	err := q.QueryRow(ctx,
		`SELECT id, user_id, code_hash, expires_at, attempts, created_at
		 FROM password_reset_codes
		 WHERE user_id=$1 AND used_at IS NULL AND expires_at > now() AND attempts < 5
		 ORDER BY created_at DESC
		 LIMIT 1`, userID,
	).Scan(&r.ID, &r.UserID, &r.CodeHash, &r.ExpiresAt, &r.Attempts, &r.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return PasswordResetCode{}, ErrNotFound
	}
	return r, err
}

func IncrementPasswordResetAttempts(ctx context.Context, q db.Conn, id uuid.UUID) error {
	_, err := q.Exec(ctx, `UPDATE password_reset_codes SET attempts=attempts+1 WHERE id=$1`, id)
	return err
}

func UsePasswordResetCode(ctx context.Context, q db.Conn, id uuid.UUID) error {
	_, err := q.Exec(ctx, `UPDATE password_reset_codes SET used_at=now() WHERE id=$1 AND used_at IS NULL`, id)
	return err
}

func InvalidatePasswordResetCodes(ctx context.Context, q db.Conn, userID uuid.UUID) error {
	_, err := q.Exec(ctx, `UPDATE password_reset_codes SET used_at=now() WHERE user_id=$1 AND used_at IS NULL`, userID)
	return err
}

// ---------- Posts ----------

const postCols = `id, slug, title, excerpt, excerpt_source, excerpt_reviewed_body_hash, body_md, body_html, cover_url, section, status, commit_hash, read_minutes, words, pinned, published_at, created_at, updated_at`

func scanPost(row pgx.Row, p *model.Post) error {
	return row.Scan(
		&p.ID, &p.Slug, &p.Title, &p.Excerpt, &p.ExcerptSource, &p.ExcerptReviewedBodyHash, &p.BodyMD, &p.BodyHTML, &p.CoverURL,
		&p.Section, &p.Status, &p.CommitHash, &p.ReadMinutes, &p.Words, &p.Pinned,
		&p.PublishedAt, &p.CreatedAt, &p.UpdatedAt,
	)
}

// ListPublished returns recent published posts (with tags).
func ListPublished(ctx context.Context, q db.Conn, limit int) ([]model.Post, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := q.Query(ctx,
		`SELECT `+postCols+` FROM posts
		 WHERE status='published'
		 ORDER BY pinned DESC NULLS LAST, published_at DESC NULLS LAST, created_at DESC
		 LIMIT $1`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectPosts(ctx, q, rows)
}

// ListPublishedAll returns every published post from newest to oldest.
func ListPublishedAll(ctx context.Context, q db.Conn) ([]model.Post, error) {
	rows, err := q.Query(ctx,
		`SELECT `+postCols+` FROM posts
		 WHERE status='published'
		 ORDER BY published_at DESC NULLS LAST, created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectPosts(ctx, q, rows)
}

// CountPublished returns the true number of published posts.
func CountPublished(ctx context.Context, q db.Conn) (int, error) {
	var count int
	err := q.QueryRow(ctx, `SELECT COUNT(*) FROM posts WHERE status='published'`).Scan(&count)
	return count, err
}

// SumPublishedWords returns the total word count across all published posts.
func SumPublishedWords(ctx context.Context, q db.Conn) (int, error) {
	var sum int
	err := q.QueryRow(ctx, `SELECT COALESCE(SUM(words), 0) FROM posts WHERE status='published'`).Scan(&sum)
	return sum, err
}
func ListAll(ctx context.Context, q db.Conn) ([]model.Post, error) {
	rows, err := q.Query(ctx,
		`SELECT `+postCols+` FROM posts ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectPosts(ctx, q, rows)
}

func GetPostBySlug(ctx context.Context, q db.Conn, slug string) (model.Post, error) {
	var p model.Post
	err := q.QueryRow(ctx,
		`SELECT `+postCols+` FROM posts WHERE slug=$1`, slug,
	).Scan(&p.ID, &p.Slug, &p.Title, &p.Excerpt, &p.ExcerptSource, &p.ExcerptReviewedBodyHash, &p.BodyMD, &p.BodyHTML, &p.CoverURL,
		&p.Section, &p.Status, &p.CommitHash, &p.ReadMinutes, &p.Words, &p.Pinned,
		&p.PublishedAt, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Post{}, ErrNotFound
	}
	if err != nil {
		return model.Post{}, err
	}
	p.Tags, _ = TagsForPost(ctx, q, p.ID)
	return p, nil
}

func GetPostByID(ctx context.Context, q db.Conn, id uuid.UUID) (model.Post, error) {
	var p model.Post
	err := q.QueryRow(ctx,
		`SELECT `+postCols+` FROM posts WHERE id=$1`, id,
	).Scan(&p.ID, &p.Slug, &p.Title, &p.Excerpt, &p.ExcerptSource, &p.ExcerptReviewedBodyHash, &p.BodyMD, &p.BodyHTML, &p.CoverURL,
		&p.Section, &p.Status, &p.CommitHash, &p.ReadMinutes, &p.Words, &p.Pinned,
		&p.PublishedAt, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Post{}, ErrNotFound
	}
	if err != nil {
		return model.Post{}, err
	}
	p.Tags, _ = TagsForPost(ctx, q, p.ID)
	return p, nil
}

func collectPosts(ctx context.Context, pool db.Conn, rows pgx.Rows) ([]model.Post, error) {
	var out []model.Post
	for rows.Next() {
		var p model.Post
		if err := scanPost(rows, &p); err != nil {
			return nil, err
		}
		p.Tags, _ = TagsForPost(ctx, pool, p.ID)
		out = append(out, p)
	}
	return out, rows.Err()
}

func PinnedPost(ctx context.Context, q db.Conn) (*model.Post, error) {
	var p model.Post
	err := q.QueryRow(ctx,
		`SELECT `+postCols+` FROM posts WHERE pinned=true AND status='published' LIMIT 1`,
	).Scan(&p.ID, &p.Slug, &p.Title, &p.Excerpt, &p.ExcerptSource, &p.ExcerptReviewedBodyHash, &p.BodyMD, &p.BodyHTML, &p.CoverURL,
		&p.Section, &p.Status, &p.CommitHash, &p.ReadMinutes, &p.Words, &p.Pinned,
		&p.PublishedAt, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p.Tags, _ = TagsForPost(ctx, q, p.ID)
	return &p, nil
}

type PostInput struct {
	Slug                    string   `json:"slug"`
	Title                   string   `json:"title"`
	Excerpt                 string   `json:"excerpt"`
	ExcerptSource           string   `json:"excerpt_source"`
	ExcerptReviewedBodyHash string   `json:"excerpt_reviewed_body_hash"`
	BodyMD                  string   `json:"body_md"`
	BodyHTML                string   `json:"body_html"`
	CoverURL                string   `json:"cover_url"`
	Section                 string   `json:"section"`
	Status                  string   `json:"status"`
	Pinned                  bool     `json:"pinned"`
	TagNames                []string `json:"tag_names"`
}

func CreatePost(ctx context.Context, q db.Conn, in PostInput, commit string, readMin, words int) (model.Post, error) {
	now := time.Now()
	var publishedAt *time.Time
	if in.Status == "published" {
		publishedAt = &now
	}
	var p model.Post
	err := q.QueryRow(ctx,
		`INSERT INTO posts (slug,title,excerpt,excerpt_source,excerpt_reviewed_body_hash,body_md,body_html,cover_url,section,status,commit_hash,read_minutes,words,pinned,published_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		 RETURNING `+postCols,
		in.Slug, in.Title, in.Excerpt, in.ExcerptSource, in.ExcerptReviewedBodyHash, in.BodyMD, in.BodyHTML, in.CoverURL, in.Section,
		in.Status, commit, readMin, words, in.Pinned, publishedAt,
	).Scan(&p.ID, &p.Slug, &p.Title, &p.Excerpt, &p.ExcerptSource, &p.ExcerptReviewedBodyHash, &p.BodyMD, &p.BodyHTML, &p.CoverURL,
		&p.Section, &p.Status, &p.CommitHash, &p.ReadMinutes, &p.Words, &p.Pinned,
		&p.PublishedAt, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return model.Post{}, err
	}
	if err := replacePostTags(ctx, q, p.ID, in.TagNames); err != nil {
		return model.Post{}, err
	}
	p.Tags, _ = TagsForPost(ctx, q, p.ID)
	return p, nil
}

func UpdatePost(ctx context.Context, q db.Conn, id uuid.UUID, in PostInput, commit string, readMin, words int) (model.Post, error) {
	now := time.Now()
	var publishedAt *time.Time
	// fetch existing published_at if any
	var existing model.Post
	_ = q.QueryRow(ctx, `SELECT published_at, status FROM posts WHERE id=$1`, id).Scan(&existing.PublishedAt, &existing.Status)
	publishedAt = existing.PublishedAt
	if in.Status == "published" && publishedAt == nil {
		publishedAt = &now
	}

	var p model.Post
	err := q.QueryRow(ctx,
		`UPDATE posts SET slug=$2,title=$3,excerpt=$4,excerpt_source=$5,excerpt_reviewed_body_hash=$6,body_md=$7,body_html=$8,cover_url=$9,section=$10,status=$11,commit_hash=$12,read_minutes=$13,words=$14,pinned=$15,published_at=$16,updated_at=now()
		 WHERE id=$1 RETURNING `+postCols,
		id, in.Slug, in.Title, in.Excerpt, in.ExcerptSource, in.ExcerptReviewedBodyHash, in.BodyMD, in.BodyHTML, in.CoverURL, in.Section,
		in.Status, commit, readMin, words, in.Pinned, publishedAt,
	).Scan(&p.ID, &p.Slug, &p.Title, &p.Excerpt, &p.ExcerptSource, &p.ExcerptReviewedBodyHash, &p.BodyMD, &p.BodyHTML, &p.CoverURL,
		&p.Section, &p.Status, &p.CommitHash, &p.ReadMinutes, &p.Words, &p.Pinned,
		&p.PublishedAt, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Post{}, ErrNotFound
	}
	if err != nil {
		return model.Post{}, err
	}
	if err := replacePostTags(ctx, q, p.ID, in.TagNames); err != nil {
		return model.Post{}, err
	}
	p.Tags, _ = TagsForPost(ctx, q, p.ID)
	return p, nil
}

func DeletePost(ctx context.Context, q db.Conn, id uuid.UUID) error {
	ct, err := q.Exec(ctx, `DELETE FROM posts WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ---------- Tags ----------

func TagsForPost(ctx context.Context, q db.Conn, postID uuid.UUID) ([]model.Tag, error) {
	rows, err := q.Query(ctx,
		`SELECT t.id, t.slug, t.name FROM tags t
		 JOIN post_tags pt ON pt.tag_id=t.id WHERE pt.post_id=$1 ORDER BY t.name`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]model.Tag, 0)
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.Slug, &t.Name); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func replacePostTags(ctx context.Context, q db.Conn, postID uuid.UUID, names []string) error {
	if _, err := q.Exec(ctx, `DELETE FROM post_tags WHERE post_id=$1`, postID); err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, name := range names {
		name = trim(name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		slug := slugify(name)
		var tag model.Tag
		err := q.QueryRow(ctx,
			`INSERT INTO tags (slug,name) VALUES ($1,$2)
			 ON CONFLICT (slug) DO UPDATE SET name=EXCLUDED.name
			 RETURNING id, slug, name`, slug, name,
		).Scan(&tag.ID, &tag.Slug, &tag.Name)
		if err != nil {
			return err
		}
		if _, err := q.Exec(ctx,
			`INSERT INTO post_tags (post_id, tag_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
			postID, tag.ID); err != nil {
			return err
		}
	}
	return nil
}

func ListPublishedArchive(ctx context.Context, q db.Conn) ([]model.Post, error) {
	rows, err := q.Query(ctx,
		`SELECT `+postCols+` FROM posts
		 WHERE status='published'
		 ORDER BY published_at DESC NULLS LAST, created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectPosts(ctx, q, rows)
}

func ListPublishedBySection(ctx context.Context, q db.Conn, section string) ([]model.Post, error) {
	rows, err := q.Query(ctx,
		`SELECT `+postCols+` FROM posts
		 WHERE status='published' AND section=$1
		 ORDER BY published_at DESC NULLS LAST, created_at DESC`, section,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectPosts(ctx, q, rows)
}

func ListPublishedByTag(ctx context.Context, q db.Conn, slug string) ([]model.Post, error) {
	rows, err := q.Query(ctx,
		`SELECT p.id, p.slug, p.title, p.excerpt, p.excerpt_source, p.excerpt_reviewed_body_hash, p.body_md, p.body_html, p.cover_url, p.section, p.status, p.commit_hash, p.read_minutes, p.words, p.pinned, p.published_at, p.created_at, p.updated_at FROM posts p
		 JOIN post_tags pt ON pt.post_id=p.id
		 JOIN tags t ON t.id=pt.tag_id
		 WHERE p.status='published' AND t.slug=$1
		 ORDER BY p.published_at DESC NULLS LAST, p.created_at DESC`, slug,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectPosts(ctx, q, rows)
}
func TagCounts(ctx context.Context, q db.Conn) ([]model.TagCount, error) {
	rows, err := q.Query(ctx,
		`SELECT t.slug, t.name, count(*)::int
		 FROM tags t
		 JOIN post_tags pt ON pt.tag_id=t.id
		 JOIN posts p ON p.id=pt.post_id
		 WHERE p.status='published'
		 GROUP BY t.slug, t.name
		 ORDER BY count(*) DESC, t.name ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.TagCount
	for rows.Next() {
		var item model.TagCount
		if err := rows.Scan(&item.Slug, &item.Name, &item.Count); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func SectionCounts(ctx context.Context, q db.Conn) ([]model.SectionCount, error) {
	rows, err := q.Query(ctx,
		`SELECT section, section, count(*)::int
		 FROM posts
		 WHERE status='published'
		 GROUP BY section
		 ORDER BY count(*) DESC, section ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.SectionCount
	for rows.Next() {
		var item model.SectionCount
		if err := rows.Scan(&item.Slug, &item.Name, &item.Count); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// ---------- Settings ----------

func GetSetting(ctx context.Context, q db.Conn, key string, dest any) error {
	var raw []byte
	err := q.QueryRow(ctx, `SELECT value::text FROM settings WHERE section_key=$1`, key).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, dest)
}

func GetSettingsMap(ctx context.Context, q db.Conn, keys ...string) (map[string]json.RawMessage, error) {
	out := make(map[string]json.RawMessage, len(keys))
	for _, k := range keys {
		var raw []byte
		err := q.QueryRow(ctx, `SELECT value::text FROM settings WHERE section_key=$1`, k).Scan(&raw)
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("setting %s: %w", k, err)
		}
		out[k] = append([]byte(nil), raw...)
	}
	return out, nil
}

func SetSetting(ctx context.Context, q db.Conn, key string, value any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = q.Exec(ctx,
		`INSERT INTO settings (section_key, value, updated_at) VALUES ($1,$2::jsonb,now())
		 ON CONFLICT (section_key) DO UPDATE SET value=EXCLUDED.value, updated_at=now()`,
		key, string(b))
	return err
}

func ListSettingKeys(ctx context.Context, q db.Conn) ([]string, error) {
	rows, err := q.Query(ctx, `SELECT section_key FROM settings ORDER BY section_key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

// ---------- helpers ----------

func trim(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}

func slugify(s string) string {
	out := make([]byte, 0, len(s))
	prevDash := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
			out = append(out, c)
			prevDash = false
		case c >= 'A' && c <= 'Z':
			out = append(out, c+32)
			prevDash = false
		case c >= 0x80:
			out = append(out, c)
			prevDash = false
		default:
			if !prevDash && len(out) > 0 {
				out = append(out, '-')
				prevDash = true
			}
		}
	}
	for len(out) > 0 && out[len(out)-1] == '-' {
		out = out[:len(out)-1]
	}
	if len(out) == 0 {
		return "tag"
	}
	return string(out)
}
