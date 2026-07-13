# Lancer Identity Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the public blog as an unmistakably Lancer-owned personal cockpit that combines racing precision, entry-fragger aggression, clutch calm, and cinematic optimism while preserving a quiet article-reading surface.

**Architecture:** Keep the existing Go templates, settings-backed content, and repository APIs. Add a scoped identity layer in `web/static/lancer.css`, update the shared shell and public templates to use Lancer-specific semantic classes, and ship a new idempotent migration that updates already-initialized settings rows. Preserve the current article DOM and motion logic wherever possible, then visually integrate it through scoped overrides.

**Tech Stack:** Go 1.x, Gin, `html/template`, PostgreSQL migrations, embedded static assets, HTML/CSS/vanilla JavaScript, Go contract tests.

## Global Constraints

- Follow the root `AGENTS.md` identity, privacy, copy, motion, and accessibility rules.
- Public identity is `Lancer`; do not introduce legal name, employer, school, location, personal photos, celebrity photos, team logos, or copyrighted game artwork.
- Primary surface remains light and editorial; use cinematic dark contrast only for major page openings.
- Signature motion is the Lancer Line. Do not add cursor followers, scroll-jacking, autoplay audio, fake loading screens, or constant high-frequency effects.
- Keep blank excerpts blank and conditional in every public list/detail view.
- Preserve the article paper, GitHub heatmap, reduced-motion behavior, and fast-scroll stability.
- Existing settings-backed deployments must receive the new copy through an idempotent migration.

## Design system

- Color: Grid Black `#0B0D12`, Paper White `#F5F4EF`, Clutch White `#FFFFFF`, Attack Orange `#FF4D2E`, Telemetry Cyan `#67D8FF`, Trophy Gold `#D4A94F`.
- Type: `Barlow Condensed` for poster/display roles, `Noto Sans SC` for Chinese reading, `IBM Plex Mono` for telemetry and commit metadata.
- Layout signature: cinematic chapter opener followed by precise light-space content.
- Motion signature: a single trajectory progresses through four nodes — `EXPLORE`, `BREAK`, `CLUTCH`, `NEXT`.

## Wireframes

```text
HOME
┌──────────────────────────────────────────────────────────────┐
│ L/  FIELD NOTES                POSTS  ABOUT  ARCHIVE  LOADOUT│
├──────────────────────────────────────────────────────────────┤
│ PERSONAL LOG / SEASON 26                    ┌─ EXPLORE ───┐  │
│ LANCER                                     ╱             │  │
│ 撕开防线，也守住残局。                BREAK ── CLUTCH ─ NEXT │
│ [进入日志] [现在的我]                      └──────────────┘  │
│ FOCUS / ...   MODE / ...   STATUS / ...                      │
└──────────────────────────────────────────────────────────────┘
┌──────────── CURRENT PLAY / pinned article ───────────────────┐
└──────────────────────────────────────────────────────────────┘
┌──────── FIELD NOTES ────────┬────── INFLUENCE DECK ─────────┐
│ recent article grid         │ PADDOCK / COURT / SERVER / MIX │
└─────────────────────────────┴────────────────────────────────┘
```

```text
ARTICLE
┌──────── cinematic, compact article chapter header ───────────┐
│ ROUND / date / commit       title and optional manual summary │
└──────────────────────────────────────────────────────────────┘
       ┌──────────── quiet Clutch White reading paper ─────┐
       │ body, code, quotations, headings                  │
       └───────────────────────────────────────────────────┘
       GitHub training record (only when API data exists)
```

---

### Task 1: Lock identity and add failing public contracts

**Files:**
- Existing: `AGENTS.md`
- Create: `web/lancer_identity_contract_test.go`

**Interfaces:**
- Consumes: the identity rules in root `AGENTS.md`.
- Produces: source-level contracts for the new shell, home cockpit, Lancer Line, secondary-page language, and article calm zone.

- [ ] **Step 1: Add the failing contract tests**

```go
func TestPublicShellLoadsLancerIdentityLayer(t *testing.T) {
    source := mustRead(t, "templates/layout.tmpl")
    for _, want := range []string{`/static/lancer.css`, `class="lancer-site`, `L/`, `FIELD NOTES`} {
        if !strings.Contains(source, want) { t.Errorf("shell missing %q", want) }
    }
}

func TestHomeUsesLancerCockpitAndTrajectory(t *testing.T) {
    source := mustRead(t, "templates/index.tmpl")
    for _, want := range []string{`class="lancer-hero`, `class="lancer-line`, `EXPLORE`, `BREAK`, `CLUTCH`, `NEXT`, `CURRENT PLAY`, `INFLUENCE DECK`} {
        if !strings.Contains(source, want) { t.Errorf("home missing %q", want) }
    }
}

func TestIdentityCSSDefinesRequiredTokensAndReducedMotion(t *testing.T) {
    source := mustRead(t, "static/lancer.css")
    for _, want := range []string{`--grid-black: #0B0D12`, `--attack-orange: #FF4D2E`, `--telemetry-cyan: #67D8FF`, `@media (prefers-reduced-motion: reduce)`} {
        if !strings.Contains(source, want) { t.Errorf("identity CSS missing %q", want) }
    }
}
```

- [ ] **Step 2: Run the contracts and verify RED**

Run: `go test ./web -run Lancer -count=1`

Expected: FAIL because `lancer.css` and the new semantic markup do not exist.

### Task 2: Build the shared Lancer shell and design foundation

**Files:**
- Modify: `web/templates/layout.tmpl`
- Create: `web/static/lancer.css`

**Interfaces:**
- Consumes: existing `.Site.Brand`, `.Site.Nav`, `.Site.Footer`, and `.Page` template data.
- Produces: `.lancer-site`, `.lancer-nav`, `.lancer-footer`, token definitions, typography roles, focus styles, reveal behavior, and reduced-motion safeguards for every public page.

- [ ] **Step 1: Update the shell markup**

Load Barlow Condensed, Noto Sans SC, and IBM Plex Mono, then load `/static/lancer.css` after the legacy stylesheet. Use this body and brand contract:

```html
<body class="lancer-site page-{{.Page}}">
<a class="lancer-brand" href="/" aria-label="Lancer home">
  <span class="lancer-mark">L/</span>
  <span><strong>{{.Site.Brand.Brand}}</strong><small>FIELD NOTES</small></span>
</a>
```

- [ ] **Step 2: Add the scoped foundation CSS**

```css
.lancer-site {
  --grid-black: #0B0D12;
  --paper-white: #F5F4EF;
  --clutch-white: #FFFFFF;
  --attack-orange: #FF4D2E;
  --telemetry-cyan: #67D8FF;
  --trophy-gold: #D4A94F;
  background: var(--paper-white);
  color: var(--grid-black);
  font-family: "Noto Sans SC", sans-serif;
}
.lancer-site :focus-visible { outline: 2px solid var(--attack-orange); outline-offset: 4px; }
@media (prefers-reduced-motion: reduce) {
  .lancer-site *, .lancer-site *::before, .lancer-site *::after { animation: none !important; transition-duration: .01ms !important; }
}
```

- [ ] **Step 3: Run the shell contract**

Run: `go test ./web -run 'PublicShell|IdentityCSS' -count=1`

Expected: PASS.

### Task 3: Rebuild home as the Lancer cockpit

**Files:**
- Modify: `web/templates/index.tmpl`
- Modify: `web/static/lancer.css`
- Test: `web/lancer_identity_contract_test.go`

**Interfaces:**
- Consumes: `IndexData.Hero`, `.Pinned`, `.Posts`, `.PostCount`, and `.Stack.Cells`.
- Produces: hero cockpit, inline SVG Lancer Line, current-play drawer, field-note grid, and influence deck.

- [ ] **Step 1: Implement the hero trajectory**

Use an accessible decorative SVG with one continuous path and four semantic node labels:

```html
<svg class="lancer-line" viewBox="0 0 620 360" aria-hidden="true">
  <path class="lancer-line-grid" d="M54 284 C150 282 142 174 246 176 S360 74 452 116 S516 220 576 64"/>
  <path class="lancer-line-live" d="M54 284 C150 282 142 174 246 176 S360 74 452 116 S516 220 576 64"/>
  <g class="lancer-node node-explore"><circle cx="54" cy="284" r="7"/><text x="54" y="318">EXPLORE</text></g>
  <g class="lancer-node node-break"><circle cx="246" cy="176" r="7"/><text x="246" y="210">BREAK</text></g>
  <g class="lancer-node node-clutch"><circle cx="452" cy="116" r="7"/><text x="452" y="150">CLUTCH</text></g>
  <g class="lancer-node node-next"><circle cx="576" cy="64" r="7"/><text x="576" y="98">NEXT</text></g>
</svg>
```

- [ ] **Step 2: Implement content cards and drawer motion**

Pinned content uses `CURRENT PLAY`; recent content uses `FIELD NOTES`; the four settings-backed stack cells form `INFLUENCE DECK`. Card secondary information moves within the card on hover/focus without changing layout height.

- [ ] **Step 3: Run the home contract**

Run: `go test ./web -run HomeUsesLancer -count=1`

Expected: PASS.

### Task 4: Reframe About, Archive, and Shelf as personal chapters

**Files:**
- Modify: `web/templates/about.tmpl`
- Modify: `web/templates/archive.tmpl`
- Modify: `web/templates/shelf.tmpl`
- Modify: `web/static/lancer.css`
- Test: `web/lancer_identity_contract_test.go`

**Interfaces:**
- Consumes: existing `AboutPageData`, `ArchivePageData`, and `ShelfPageData` only; no handler shape changes.
- Produces: `PLAYER PROFILE`, `SEASON LOG`, and `LOADOUT` page identities.

- [ ] **Step 1: Add secondary-page failing contracts**

```go
func TestSecondaryPagesUseLancerChapters(t *testing.T) {
    checks := map[string][]string{
        "templates/about.tmpl": {`PLAYER PROFILE`, `ON TRACK`, `OFF TRACK`, `CURRENT ROUND`},
        "templates/archive.tmpl": {`SEASON LOG`, `RACE CONTROL`, `class="season-line`},
        "templates/shelf.tmpl": {`LOADOUT`, `INFLUENCE DECK`, `class="loadout-card`},
    }
    for file, wants := range checks { source := mustRead(t, file); for _, want := range wants { if !strings.Contains(source, want) { t.Errorf("%s missing %q", file, want) } } }
}
```

- [ ] **Step 2: Run and verify RED**

Run: `go test ./web -run SecondaryPages -count=1`

Expected: FAIL with the new chapter labels missing.

- [ ] **Step 3: Implement the three chapter layouts**

About separates values (`ON TRACK`) from interests (`OFF TRACK`) and keeps the terminal only for `CURRENT ROUND`. Archive treats years as seasons with an orange route line. Shelf treats groups as configurable loadouts and influences, with drawer-like cards.

- [ ] **Step 4: Run and verify GREEN**

Run: `go test ./web -run SecondaryPages -count=1`

Expected: PASS.

### Task 5: Integrate the article paper and live training record

**Files:**
- Modify: `web/templates/post.tmpl`
- Modify: `web/static/lancer.css`
- Modify: `web/post_page_contract_test.go`

**Interfaces:**
- Consumes: the existing article paper, excerpt conditional, border trace script, and `.Heatmap` conditional.
- Produces: a compact `ROUND NOTES` article opener and a visually quieter reading paper under the global Lancer shell.

- [ ] **Step 1: Add the failing article identity contract**

```go
for _, want := range []string{`ROUND NOTES`, `class="article-cockpit`, `class="article-paper`, `{{if .Post.Excerpt}}`, `{{if .Heatmap}}`} {
    if !strings.Contains(source, want) { t.Errorf("article identity missing %q", want) }
}
```

- [ ] **Step 2: Run and verify RED**

Run: `go test ./web -run 'PostPage|Article' -count=1`

Expected: FAIL only for the new article identity labels/classes.

- [ ] **Step 3: Implement the article cockpit and calm-zone overrides**

Keep the title, optional excerpt, author/date/read/words/commit metadata, article body, trace behavior, and heatmap data. Reduce the surrounding visual noise once the reader enters `.article-paper`.

- [ ] **Step 4: Run and verify GREEN**

Run: `go test ./web -run 'PostPage|Article' -count=1`

Expected: PASS.

### Task 6: Ship Lancer-owned settings content to existing databases

**Files:**
- Create: `web/migrations/0004_lancer_identity.sql`
- Modify: `internal/handler/public.go`
- Test: `web/lancer_identity_contract_test.go`

**Interfaces:**
- Consumes: existing `branding`, `hero`, `stack`, `about`, `now`, `archive`, and `shelf` settings schemas.
- Produces: Chinese-first Lancer copy and influence content for both upgraded and first-run databases.

- [ ] **Step 1: Add the failing migration contract**

```go
func TestLancerIdentityMigrationUpdatesExistingSettings(t *testing.T) {
    source := mustRead(t, "migrations/0004_lancer_identity.sql")
    for _, want := range []string{`UPDATE settings`, `'hero'`, `'stack'`, `'about'`, `'now'`, `'archive'`, `'shelf'`, `Lancer`, `这一回合没结束`} {
        if !strings.Contains(source, want) { t.Errorf("identity migration missing %q", want) }
    }
}
```

- [ ] **Step 2: Run and verify RED**

Run: `go test ./web -run LancerIdentityMigration -count=1`

Expected: FAIL because migration `0004` does not exist.

- [ ] **Step 3: Implement idempotent settings updates and matching Go fallbacks**

Use direct `UPDATE settings SET value = ... WHERE section_key = ...` statements so existing deployments change, and update `defaultAbout`, `defaultNow`, `defaultArchiveIntro`, and `defaultShelf` to match the same identity when rows are absent.

- [ ] **Step 4: Run and verify GREEN**

Run: `go test ./web -run LancerIdentityMigration -count=1`

Expected: PASS.

### Task 7: Full verification, responsive visual QA, and deploy artifact

**Files:**
- Modify only if QA exposes a defect in files already listed above.

**Interfaces:**
- Consumes: all previous tasks.
- Produces: validated desktop/mobile public pages and an embedded Linux deployment binary.

- [ ] **Step 1: Run automated verification**

```powershell
go test ./... -count=1
go run ./cmd/tplcheck
git diff --check
```

Expected: all commands exit `0`; every template parses.

- [ ] **Step 2: Start local development server**

Run from `web/admin`: `npm run dev:local`

Expected: local public site and admin start without migration/template errors.

- [ ] **Step 3: Perform visual QA**

Inspect `/`, `/about`, `/archive`, `/shelf`, and one `/posts/:slug` page at desktop and mobile widths. Verify typography, route motion, drawer focus/hover, excerpt absence, article calm, heatmap condition, fast scrolling, and reduced-motion fallback.

- [ ] **Step 4: Build the deploy artifact**

Run from `web/admin`: `npm run build:deploy`

Expected: `web/admin-dist` builds and `bin/blog` starts with ELF magic `7F 45 4C 46`.
