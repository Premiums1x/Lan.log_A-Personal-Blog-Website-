# AI Excerpt Workflow Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Generate editable article excerpts through a configurable LLM and ask whether to refresh the excerpt whenever a published article body changes.

**Architecture:** PostgreSQL stores excerpt provenance and the body hash last reviewed against that excerpt. A server-only OpenAI-compatible client generates summaries through an authenticated endpoint. The admin editor detects stale or changed bodies, asks for a decision before publishing, previews the generated copy, and never silently overwrites manual text.

**Tech Stack:** Go, Gin, PostgreSQL migrations, React 18, TypeScript, Ant Design, TanStack Query.

## Global Constraints

- Blank excerpts remain blank and never fall back to body truncation.
- LLM credentials stay server-side in environment variables.
- AI failure must not overwrite existing copy or block saving through the non-AI choices.
- Preserve current uncommitted article-motion changes.
- Do not add a new frontend dependency.

---

### Task 1: Persist excerpt review state

**Files:**
- Create: `web/migrations/0003_ai_excerpt.sql`
- Modify: `internal/model/model.go`
- Modify: `internal/repo/repo.go`
- Test: `internal/handler/api_excerpt_test.go`

**Interfaces:**
- `Post.ExcerptSource string`
- `Post.ExcerptReviewedBodyHash string`
- `Post.ExcerptStale bool`
- `PostInput.ExcerptSource string`
- `PostInput.ExcerptReviewedBodyHash string`

- [ ] Write failing tests for SHA-256 body hashing, source normalization, and stale detection.
- [ ] Run `go test ./internal/handler -run Excerpt -count=1` and confirm the new tests fail.
- [ ] Add `excerpt_source` and `excerpt_reviewed_body_hash` columns with existing rows classified as `manual` or `empty`.
- [ ] Extend post model, select/scan, insert, and update paths.
- [ ] Implement helper behavior: blank forces `empty`; `ai` is retained only for nonblank copy; otherwise source is `manual`; reviewed hash equals SHA-256 of current Markdown.
- [ ] Rerun the focused tests and confirm they pass.

### Task 2: Add the server-side LLM client

**Files:**
- Create: `internal/llm/client.go`
- Create: `internal/llm/client_test.go`
- Modify: `internal/config/config.go`
- Modify: `.env.example`

**Interfaces:**
- `config.LLMConfig { APIURL, APIKey, Model string; TimeoutSeconds int }`
- `func (c LLMConfig) Ready() bool`
- `func NewClient(config Config) *Client`
- `func (c *Client) GenerateExcerpt(ctx context.Context, title, body string) (string, error)`

- [ ] Write `httptest` tests proving authorization/model payload, response extraction, empty-response rejection, and non-2xx handling.
- [ ] Run `go test ./internal/llm -count=1` and confirm failure before implementation.
- [ ] Implement an OpenAI-compatible chat-completions request with a Chinese 80-120 character, plain-text, no-invention system instruction and a bounded timeout.
- [ ] Add `LLM_API_URL`, `LLM_API_KEY`, `LLM_MODEL`, and `LLM_TIMEOUT_SECONDS` configuration.
- [ ] Rerun the client tests and confirm they pass.

### Task 3: Expose authenticated excerpt generation

**Files:**
- Modify: `internal/handler/api.go`
- Modify: `internal/server/server.go`
- Test: `internal/handler/api_excerpt_test.go`

**Interfaces:**
- `POST /api/ai/excerpt` with `{title, body_md}`
- Response `{excerpt}`
- Post writes accept `excerpt_source` and `excerpt_reviewed`.

- [ ] Write failing helper/contract tests for request validation, stale decoration, and route registration.
- [ ] Run focused handler tests and confirm failure.
- [ ] Add authenticated route and handler; return 503 when LLM configuration is absent, 400 for empty body, and 502 for generation failure.
- [ ] On post update preserve the old reviewed hash unless the client explicitly reviewed the excerpt or changed manual excerpt text; then hash the current body.
- [ ] Return `excerpt_stale` on post detail responses.
- [ ] Rerun focused tests and confirm pass.

### Task 4: Add the admin review workflow

**Files:**
- Modify: `web/admin/src/api/client.ts`
- Modify: `web/admin/src/pages/PostEdit.tsx`
- Test: `web/post_page_contract_test.go`

**Interfaces:**
- Post fields: `excerpt_source`, `excerpt_reviewed_body_hash`, `excerpt_stale`.
- Publish review choices: generate with AI, keep current, clear excerpt, cancel.

- [ ] Add a failing contract test for the authenticated endpoint call, stale-body condition, modal copy, and review buttons.
- [ ] Run `go test ./web -run AIExcerpt -count=1` and confirm failure.
- [ ] Track original body/source, open the modal only for blank new publishes or changed/stale published articles, and leave draft saving nonblocking.
- [ ] Generate into an editable preview; publish only after explicit confirmation.
- [ ] Keep and clear actions publish immediately with `excerpt_reviewed=true`; AI failures keep the dialog and old excerpt intact.
- [ ] Run the focused contract test and `npm run build`.

### Task 5: Verify the integrated change

**Files:**
- Verify all modified files.

- [ ] Run `go test ./... -count=1`.
- [ ] Run `go run ./cmd/tplcheck`.
- [ ] Run `npm run build` in `web/admin`.
- [ ] Run `git diff --check` and inspect `git diff` for unrelated edits.