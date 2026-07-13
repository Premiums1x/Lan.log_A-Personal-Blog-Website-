# Lancer Live World Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn the four public landing pages into full-viewport Lancer openers with drawer-like chapter transitions, licensed interest imagery beyond the hero, and a live but privacy-safe server telemetry card.

**Architecture:** Keep code, terminal, commit, and telemetry as the interface dialect; use NiKo/F1/football/basketball imagery as the visual world behind and between content. Add one public read-only `/api/system-status` endpoint backed by a small Linux `/proc` collector, then render it progressively in the home drawer without making page rendering depend on telemetry availability.

**Tech Stack:** Go 1.26, Gin, `html/template`, embedded static assets, vanilla JavaScript, CSS, Linux `/proc`, Go tests.

## Global Constraints

- Public identity remains `Lancer`; do not expose hostname, IP, process names, paths, legal name, employer, or location.
- Preserve explicit blank excerpts and the quiet independent article paper.
- Use properly licensed source images only, keep visible media credits, and record any crop or color-grade transformation.
- No scroll-jacking, scroll snapping, autoplay media, cursor chasers, or high-frequency effects.
- All motion must have a `prefers-reduced-motion` fallback and remain stable during fast scrolling.
- Do not turn every section into a card; reserve cards for the interest deck and live telemetry.
- Required final verification: `go test ./... -count=1`, `go run ./cmd/tplcheck`, desktop visual QA, and 390 px mobile visual QA.

---

## File Map

- `web/static/media/interests/*`: locally served, optimized interest photographs.
- `web/static/media/interests/CREDITS.md`: source, author, license, source URL, and transformation record for every photograph.
- `web/templates/index.tmpl`: full-screen home opener, drawer wrapper, World Deck, and telemetry card.
- `web/templates/about.tmpl`: full-screen profile opener, drawer wrapper, image-backed interactive influences.
- `web/templates/archive.tmpl`: full-screen season opener and drawer wrapper.
- `web/templates/shelf.tmpl`: full-screen loadout opener and drawer wrapper.
- `web/templates/layout.tmpl`: shared media-credit disclosure and shared client script hook.
- `web/static/lancer.css`: full-screen/drawer geometry, image treatment, cards, telemetry, and reduced-motion rules.
- `web/static/lancer.js`: drawer cue, influence switching, and telemetry polling; keeps behavior out of templates.
- `internal/telemetry/collector.go`: public snapshot types and cached collector interface.
- `internal/telemetry/procfs.go`: Linux `/proc` parsers and sampling implementation.
- `internal/telemetry/procfs_test.go`: parser, delta, cache, and privacy tests.
- `internal/handler/system_status.go`: JSON handler with no-store response.
- `internal/handler/system_status_test.go`: success and unavailable response contract tests.
- `internal/server/server.go`: registers the public endpoint before the authenticated API group.
- `web/lancer_identity_contract_test.go`: template/static contracts for drawers, imagery, cards, and reduced motion.

---

### Task 1: License and install the interest image set

**Files:**
- Create: `web/static/media/interests/niko-2022.webp`
- Create: `web/static/media/interests/verstappen-2018.webp`
- Create: `web/static/media/interests/leclerc.webp`
- Create: `web/static/media/interests/messi.webp`
- Create: `web/static/media/interests/CREDITS.md`
- Test: `web/lancer_identity_contract_test.go`

**Interfaces:**
- Consumes: Wikimedia Commons source pages and licenses listed below.
- Produces: four local WebP paths used by CSS custom properties and templates.

- [ ] **Step 1: Add a failing asset-and-credit contract**

```go
func TestInterestImagesAreLocalAndCredited(t *testing.T) {
	assets := []string{"niko-2022.webp", "verstappen-2018.webp", "leclerc.webp", "messi.webp"}
	credits := readIdentitySource(t, "static/media/interests/CREDITS.md")
	for _, asset := range assets {
		if _, err := os.Stat("static/media/interests/" + asset); err != nil {
			t.Errorf("interest asset %s missing: %v", asset, err)
		}
		requireIdentityStrings(t, credits, "media credits", asset)
	}
	requireIdentityStrings(t, credits, "media credits", "CC BY 4.0", "CC0 1.0", "CC BY-SA 4.0", "CC BY 2.0")
}
```

- [ ] **Step 2: Run the contract to verify it fails**

Run: `go test ./web -run TestInterestImagesAreLocalAndCredited -count=1`

Expected: FAIL because the interest files and credit manifest do not exist.

- [ ] **Step 3: Download and optimize the licensed sources**

Use these exact Commons records:

```text
NiKo 2022 — G2 Esports — CC BY 4.0
https://commons.wikimedia.org/wiki/File:NiKo_2022.png

Max Verstappen — Fotóshírek szerkesztőség; description credit © Práger Péter / www.pragerfoto.hu — CC0 1.0
https://commons.wikimedia.org/wiki/File:Max_Verstappen_-_Nagy_Futam_2018.jpg

Charles Leclerc — Gil Zetbase — CC BY-SA 4.0
https://commons.wikimedia.org/wiki/File:Charles-Leclerc.jpg

Lionel Messi — Wonker — CC BY 2.0
https://commons.wikimedia.org/wiki/File:Lionel_Messi.jpg
```

Download via `Special:Redirect/file/...`, crop for responsive background use, convert to WebP at quality 82, and cap the longest edge at 1800 px. Do not hotlink. Record `cropped, resized, monochrome/color grade` beside every derivative in `CREDITS.md`.

- [ ] **Step 4: Add visible credit data**

The manifest must contain this schema for every asset:

```markdown
| Local asset | Subject | Author | License | Source | Changes |
| --- | --- | --- | --- | --- | --- |
| `niko-2022.webp` | NiKo | G2 Esports | [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/) | [Wikimedia Commons](https://commons.wikimedia.org/wiki/File:NiKo_2022.png) | Cropped, resized, and color graded. |
```

- [ ] **Step 5: Run the contract**

Run: `go test ./web -run TestInterestImagesAreLocalAndCredited -count=1`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add web/static/media/interests web/lancer_identity_contract_test.go
git commit -m "feat: add licensed Lancer influence media"
```

---

### Task 2: Build the full-screen opener and drawer transition

**Files:**
- Modify: `web/templates/index.tmpl`
- Modify: `web/templates/about.tmpl`
- Modify: `web/templates/archive.tmpl`
- Modify: `web/templates/shelf.tmpl`
- Modify: `web/static/lancer.css`
- Modify: `web/static/lancer.js`
- Test: `web/lancer_identity_contract_test.go`

**Interfaces:**
- Consumes: existing `.cinematic-stage` hero markup.
- Produces: `.page-opening`, `.drawer-cue`, `.page-drawer`, and `data-drawer-target` on all four primary pages.

- [ ] **Step 1: Add failing opener contracts**

```go
func TestPrimaryPagesUseViewportOpenersAndDrawerShells(t *testing.T) {
	for _, file := range []string{"templates/index.tmpl", "templates/about.tmpl", "templates/archive.tmpl", "templates/shelf.tmpl"} {
		source := readIdentitySource(t, file)
		requireIdentityStrings(t, source, file, `page-opening`, `class="drawer-cue`, `class="page-drawer`, `data-drawer-target`)
	}
	css := readIdentitySource(t, "static/lancer.css")
	requireIdentityStrings(t, css, "drawer CSS", `min-height: calc(100svh`, `.page-drawer`, `@media (prefers-reduced-motion: reduce)`)
}
```

- [ ] **Step 2: Run the test and confirm failure**

Run: `go test ./web -run TestPrimaryPagesUseViewportOpenersAndDrawerShells -count=1`

Expected: FAIL with missing `page-opening`.

- [ ] **Step 3: Update all four template shells**

Use this exact semantic pattern, with a unique target on each page:

```html
<header class="... page-opening">
  <!-- existing hero content -->
  <a class="drawer-cue" href="#home-drawer" data-drawer-target="home-drawer" aria-label="继续向下浏览">
    <span>SCROLL TO OPEN</span><b aria-hidden="true">↓</b>
  </a>
</header>
<div class="page-drawer" id="home-drawer">
  <!-- all existing sections after the hero -->
</div>
```

Use `home-drawer`, `about-drawer`, `archive-drawer`, and `shelf-drawer`. Do not duplicate an ID and do not hide the target from keyboard users.

- [ ] **Step 4: Implement stable drawer geometry**

Add the following foundation and then adapt existing page-specific spacing:

```css
.page-opening {
  min-height: calc(100svh - var(--lancer-nav-height, 76px));
  display: grid;
  align-content: center;
  isolation: isolate;
}

.page-drawer {
  position: relative;
  z-index: 4;
  margin-top: -2rem;
  border-radius: 2rem 2rem 0 0;
  background: var(--paper-white);
  box-shadow: 0 -24px 70px rgba(11, 13, 18, .18);
  overflow: clip;
}

.drawer-cue {
  position: absolute;
  left: 50%;
  bottom: 2rem;
  transform: translateX(-50%);
  display: grid;
  justify-items: center;
  gap: .45rem;
}

.drawer-cue b { animation: drawer-cue-step 1.8s steps(2, end) infinite; }
@keyframes drawer-cue-step { 50% { transform: translateY(.45rem); } }
```

The drawer moves through normal document flow; do not bind transforms to every scroll event and do not add `scroll-snap-type`.

- [ ] **Step 5: Add progressive cue behavior**

`web/static/lancer.js` must use native anchor navigation and only mark the cue as seen:

```js
document.querySelectorAll('[data-drawer-target]').forEach((cue) => {
  cue.addEventListener('click', () => cue.classList.add('is-used'), { once: true });
});
```

- [ ] **Step 6: Add reduced-motion and mobile rules**

```css
@media (max-width: 640px) {
  .page-opening { min-height: calc(100svh - 68px); }
  .page-drawer { margin-top: -1rem; border-radius: 1.25rem 1.25rem 0 0; }
}

@media (prefers-reduced-motion: reduce) {
  .drawer-cue b { animation: none; }
  html { scroll-behavior: auto; }
}
```

- [ ] **Step 7: Verify and commit**

Run: `go test ./web -run TestPrimaryPagesUseViewportOpenersAndDrawerShells -count=1`

Expected: PASS.

```bash
git add web/templates web/static/lancer.css web/static/lancer.js web/lancer_identity_contract_test.go
git commit -m "feat: add full-screen drawer page openings"
```

---

### Task 3: Place the interest world beyond the first component

**Files:**
- Modify: `web/templates/index.tmpl`
- Modify: `web/templates/about.tmpl`
- Modify: `web/templates/archive.tmpl`
- Modify: `web/templates/shelf.tmpl`
- Modify: `web/templates/layout.tmpl`
- Modify: `web/static/lancer.css`
- Modify: `web/static/lancer.js`
- Test: `web/lancer_identity_contract_test.go`

**Interfaces:**
- Consumes: local image files from Task 1 and drawer shells from Task 2.
- Produces: `.world-deck`, `.world-card`, `.chapter-image-band`, image-backed About influence states, and visible media credits.

- [ ] **Step 1: Add failing visual-world contracts**

```go
func TestInterestWorldAppearsBeyondEveryHero(t *testing.T) {
	checks := map[string][]string{
		"templates/index.tmpl": {`class="world-deck`, `data-world="niko"`, `class="chapter-image-band`},
		"templates/about.tmpl": {`--influence-image`, `niko-2022.webp`, `verstappen-2018.webp`},
		"templates/archive.tmpl": {`class="chapter-image-band`, `leclerc.webp`},
		"templates/shelf.tmpl": {`class="world-card`, `messi.webp`},
		"templates/layout.tmpl": {`MEDIA CREDITS`, `Wikimedia Commons`},
	}
	for file, wants := range checks {
		requireIdentityStrings(t, readIdentitySource(t, file), file, wants...)
	}
}
```

- [ ] **Step 2: Run the contract and confirm failure**

Run: `go test ./web -run TestInterestWorldAppearsBeyondEveryHero -count=1`

Expected: FAIL with missing `world-deck`.

- [ ] **Step 3: Replace code-photo hero backgrounds with interest imagery**

Set page-specific CSS variables on the opener rather than hard-coding remote URLs:

```html
<header class="lancer-hero cinematic-stage page-opening" style="--opening-image:url('/static/media/interests/niko-2022.webp')">
```

Keep the editor/terminal panes as foreground identity layers. Apply a dark gradient and `background-position` per subject so text never crosses a face.

- [ ] **Step 4: Add the home World Deck after Current Play**

Use four real content cards, not logo tiles:

```html
<section class="world-deck" aria-labelledby="world-deck-title">
  <header><span>02 / WORLD DECK</span><h2 id="world-deck-title">让我保持进攻的东西。</h2></header>
  <div class="world-card-grid">
    <article class="world-card is-active" data-world="niko" style="--world-image:url('/static/media/interests/niko-2022.webp')" tabindex="0">
      <span>ENTRY / CLUTCH</span><h3>NiKo</h3><p>突破时相信枪法，残局里相信判断。</p>
    </article>
    <article class="world-card" data-world="f1" style="--world-image:url('/static/media/interests/verstappen-2018.webp')" tabindex="0">
      <span>RACE / PRECISION</span><h3>F1</h3><p>把速度变成一条可以重复走对的线。</p>
    </article>
    <article class="world-card" data-world="football" style="--world-image:url('/static/media/interests/messi.webp')" tabindex="0">
      <span>VISION / BREAK</span><h3>Football</h3><p>看见缝隙，然后撕开防线。</p>
    </article>
  </div>
</section>
```

The active card expands from `1fr` to `1.65fr` on desktop; mobile uses a natural horizontal scroll row without `scroll-snap`.

- [ ] **Step 5: Add mid-page image bands to Archive and Shelf**

Add one `.chapter-image-band` between major content groups on each page. Each band must be at least `min(52rem, 76vw)` wide and `clamp(18rem, 42vw, 34rem)` tall, with text and mono telemetry layered over the lower third. This is the explicit answer to “backgrounds only in the first component”: the image becomes a page chapter, not a hero-only wallpaper.

- [ ] **Step 6: Upgrade About INFLUENCES with real image states**

Each existing control receives `data-image` and sets `--influence-image`. Move the existing inline interaction into `lancer.js`, preserving `pointerenter`, `focusin`, `click`, and `aria-selected`. The preview must crossfade only when the selected state changes; it must not chase the pointer.

- [ ] **Step 7: Add visible media credits**

Add a compact `<details class="media-credits">` to the shared footer. It must link to each Commons source and license page. The visible label is `MEDIA CREDITS`; the closed state stays visually quiet.

- [ ] **Step 8: Verify and commit**

Run: `go test ./web -run 'TestInterestWorldAppearsBeyondEveryHero|TestAboutInfluencesSwitchOnPointerAndKeyboardFocus' -count=1`

Expected: PASS.

```bash
git add web/templates web/static/lancer.css web/static/lancer.js web/lancer_identity_contract_test.go
git commit -m "feat: carry Lancer influences through public pages"
```

---

### Task 4: Build the privacy-safe Linux telemetry collector

**Files:**
- Create: `internal/telemetry/collector.go`
- Create: `internal/telemetry/procfs.go`
- Create: `internal/telemetry/procfs_test.go`

**Interfaces:**
- Produces: `telemetry.Snapshot`, `telemetry.Provider`, and `telemetry.NewProcFSProvider(root string, ttl time.Duration) Provider`.
- Consumed by: Task 5 HTTP handler.

- [ ] **Step 1: Write parser and privacy tests**

```go
func TestParseMeminfoUsesMemAvailable(t *testing.T) {
	total, used, pct, err := parseMeminfo("MemTotal: 1000 kB\nMemAvailable: 375 kB\n")
	if err != nil || total != 1024000 || used != 640000 || pct != 62.5 {
		t.Fatalf("got total=%d used=%d pct=%v err=%v", total, used, pct, err)
	}
}

func TestCPUPercentUsesCounterDelta(t *testing.T) {
	previous := cpuTimes{Total: 1000, Idle: 700}
	current := cpuTimes{Total: 1200, Idle: 760}
	if got := cpuPercent(previous, current); got != 70 {
		t.Fatalf("got %v", got)
	}
}

func TestSnapshotJSONDoesNotExposeHostIdentity(t *testing.T) {
	b, _ := json.Marshal(Snapshot{})
	for _, forbidden := range []string{"hostname", "process", "path", "ip", "mount"} {
		if strings.Contains(strings.ToLower(string(b)), forbidden) { t.Fatalf("leaked %s", forbidden) }
	}
}
```

- [ ] **Step 2: Run tests and confirm failure**

Run: `go test ./internal/telemetry -count=1`

Expected: FAIL because the package does not exist.

- [ ] **Step 3: Define the public collector contract**

```go
package telemetry

type Snapshot struct {
	Online          bool      `json:"online"`
	CPUPercent      float64   `json:"cpu_percent"`
	CPUFrequencyMHz float64   `json:"cpu_frequency_mhz"`
	MemoryUsedBytes uint64    `json:"memory_used_bytes"`
	MemoryTotalBytes uint64   `json:"memory_total_bytes"`
	MemoryPercent   float64   `json:"memory_percent"`
	UptimeSeconds   uint64    `json:"uptime_seconds"`
	SampledAt       time.Time `json:"sampled_at"`
}

type Provider interface {
	Snapshot(context.Context) (Snapshot, error)
}
```

- [ ] **Step 4: Implement `/proc` parsing and five-second caching**

`NewProcFSProvider("/proc", 5*time.Second)` reads only `stat`, `meminfo`, `cpuinfo`, and `uptime`. Guard cached state and previous CPU counters with `sync.Mutex`. Compute CPU percent from two successive snapshots; return `0` for the first sample rather than inventing a value. Average all `cpu MHz` values. Clamp all percentages to `[0, 100]`.

- [ ] **Step 5: Add fixture-based provider coverage**

Create `t.TempDir()` fixtures named `stat`, `meminfo`, `cpuinfo`, and `uptime`; assert exact JSON-facing values and assert a second call inside the TTL returns the same `SampledAt`.

- [ ] **Step 6: Run and commit**

Run: `go test ./internal/telemetry -count=1`

Expected: PASS.

```bash
git add internal/telemetry
git commit -m "feat: collect privacy-safe server telemetry"
```

---

### Task 5: Expose the public system-status endpoint

**Files:**
- Create: `internal/handler/system_status.go`
- Create: `internal/handler/system_status_test.go`
- Modify: `internal/server/server.go`

**Interfaces:**
- Consumes: `telemetry.Provider` from Task 4.
- Produces: `GET /api/system-status` with `200 Snapshot` or `503 {"online":false,"error":"telemetry unavailable"}`.

- [ ] **Step 1: Write handler contract tests with a fake provider**

```go
type fakeTelemetry struct { snapshot telemetry.Snapshot; err error }
func (f fakeTelemetry) Snapshot(context.Context) (telemetry.Snapshot, error) { return f.snapshot, f.err }

func TestSystemStatusReturnsPublicSnapshot(t *testing.T) {
	g := gin.New()
	h := NewSystemStatusHandler(fakeTelemetry{snapshot: telemetry.Snapshot{Online: true, CPUPercent: 24.5}})
	g.GET("/api/system-status", h.Show)
	r := httptest.NewRequest(http.MethodGet, "/api/system-status", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"cpu_percent":24.5`) { t.Fatalf("%d %s", w.Code, w.Body.String()) }
	if w.Header().Get("Cache-Control") != "no-store" { t.Fatalf("missing no-store") }
}
```

- [ ] **Step 2: Run the test and confirm failure**

Run: `go test ./internal/handler -run TestSystemStatus -count=1`

Expected: FAIL because `NewSystemStatusHandler` is undefined.

- [ ] **Step 3: Implement the handler**

```go
type SystemStatusHandler struct { provider telemetry.Provider }

func NewSystemStatusHandler(provider telemetry.Provider) *SystemStatusHandler {
	return &SystemStatusHandler{provider: provider}
}

func (h *SystemStatusHandler) Show(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	snapshot, err := h.provider.Snapshot(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"online": false, "error": "telemetry unavailable"})
		return
	}
	c.JSON(http.StatusOK, snapshot)
}
```

- [ ] **Step 4: Register the endpoint outside JWT middleware**

In `internal/server/server.go`:

```go
status := handler.NewSystemStatusHandler(telemetry.NewProcFSProvider("/proc", 5*time.Second))
ag := r.Group("/api")
ag.GET("/system-status", status.Show)
```

Place it beside `/api/brand`, before `authed.Use(mgr.Middleware())`.

- [ ] **Step 5: Verify and commit**

Run: `go test ./internal/handler ./internal/server -count=1`

Expected: PASS.

```bash
git add internal/handler/system_status.go internal/handler/system_status_test.go internal/server/server.go
git commit -m "feat: expose public server status snapshot"
```

---

### Task 6: Render the LIVE NODE telemetry card

**Files:**
- Modify: `web/templates/index.tmpl`
- Modify: `web/templates/layout.tmpl`
- Modify: `web/static/lancer.css`
- Modify: `web/static/lancer.js`
- Test: `web/lancer_identity_contract_test.go`

**Interfaces:**
- Consumes: `GET /api/system-status` from Task 5.
- Produces: `[data-system-status]` card with CPU, MHz, memory, and uptime outputs.

- [ ] **Step 1: Add a failing frontend contract**

```go
func TestHomeRendersLiveNodeWithoutLeakingHostIdentity(t *testing.T) {
	source := readIdentitySource(t, "templates/index.tmpl")
	requireIdentityStrings(t, source, "live node", `data-system-status`, `data-metric="cpu"`, `data-metric="memory"`, `data-metric="frequency"`, `data-metric="uptime"`)
	js := readIdentitySource(t, "static/lancer.js")
	requireIdentityStrings(t, js, "telemetry client", `/api/system-status`, `AbortController`, `visibilityState`)
	for _, forbidden := range []string{"hostname", "processes", "mounts"} {
		if strings.Contains(strings.ToLower(source+js), forbidden) { t.Fatalf("public telemetry references %s", forbidden) }
	}
}
```

- [ ] **Step 2: Run the contract and confirm failure**

Run: `go test ./web -run TestHomeRendersLiveNodeWithoutLeakingHostIdentity -count=1`

Expected: FAIL with missing `data-system-status`.

- [ ] **Step 3: Add the instrument card to the home drawer**

```html
<section class="live-node-card" data-system-status aria-labelledby="live-node-title">
  <header><span class="live-dot" aria-hidden="true"></span><div><p>LIVE NODE</p><h2 id="live-node-title">这台博客服务器，此刻正在运行。</h2></div><output data-status>CONNECTING</output></header>
  <div class="live-node-grid">
    <article><span>CPU LOAD</span><strong data-metric="cpu">--%</strong><i data-gauge="cpu"></i></article>
    <article><span>CPU CLOCK</span><strong data-metric="frequency">-- MHz</strong></article>
    <article><span>MEMORY</span><strong data-metric="memory">--%</strong><small data-memory-detail>-- / --</small><i data-gauge="memory"></i></article>
    <article><span>UPTIME</span><strong data-metric="uptime">--</strong></article>
  </div>
</section>
```

- [ ] **Step 4: Poll safely and fail visibly**

In `lancer.js`, fetch immediately and every 15 seconds only while `document.visibilityState === 'visible'`. Use an eight-second `AbortController` timeout. Update values via `textContent`, clamp gauge percentages before assigning a CSS custom property, and render `UNAVAILABLE` with blank metrics on any non-2xx response. Do not retain invented fallback numbers.

- [ ] **Step 5: Style as an instrument card**

Use one dark precision card with orange/cyan gauges, square data cells, mono labels, and a clipped top-right corner. Keep it inside the light drawer so it reads as server instrumentation, not a second hero. At 390 px, metrics become a two-column grid; at 320 px, one column.

- [ ] **Step 6: Verify and commit**

Run: `go test ./web -run TestHomeRendersLiveNodeWithoutLeakingHostIdentity -count=1`

Expected: PASS.

```bash
git add web/templates/index.tmpl web/templates/layout.tmpl web/static/lancer.css web/static/lancer.js web/lancer_identity_contract_test.go
git commit -m "feat: add live server instrument card"
```

---

### Task 7: Integrate, preview, and harden the final experience

**Files:**
- Modify: `tmp/lancer-preview/main.go` (ignored preview fixture only)
- Modify: files found during QA only when a test or visible defect proves the need.

**Interfaces:**
- Consumes: all previous tasks.
- Produces: verified desktop/mobile public pages and Linux deploy binary.

- [ ] **Step 1: Add preview telemetry fixture**

Register `/api/system-status` in the ignored preview server with deterministic changing values so frontend behavior can be inspected on Windows without `/proc`.

- [ ] **Step 2: Run complete automated verification**

```powershell
go test ./... -count=1
go run ./cmd/tplcheck
git diff --check
```

Expected: all Go packages PASS, six templates report OK, and diff check reports no whitespace errors.

- [ ] **Step 3: Run desktop visual QA at 1440x900**

Verify `/`, `/about`, `/archive`, and `/shelf`:

- first scene fills the viewport below navigation;
- cue is visible but not over primary copy;
- normal scrolling makes the light drawer cover the opener without flicker;
- favorite imagery appears in both opener and at least one later chapter;
- code/terminal foreground remains legible;
- LIVE NODE refreshes and unavailable state is honest;
- About controls switch on hover, click, and keyboard focus.

- [ ] **Step 4: Run mobile visual QA at 390x844**

Verify no horizontal page scroll, no face/text collision, cue remains reachable, drawer radius is smaller, cards are touch-scrollable, and telemetry is two columns.

- [ ] **Step 5: Check reduced motion**

Enable `prefers-reduced-motion: reduce`; verify cue animation and image crossfades stop while content and controls remain fully usable.

- [ ] **Step 6: Build the Linux artifact and verify architecture**

```powershell
npm run build:deploy --prefix web/admin
```

Expected: build succeeds and the resulting `bin/blog` begins with ELF magic and targets x86-64 Linux.

- [ ] **Step 7: Final commit**

```bash
git add -A
git commit -m "feat: deliver Lancer live world experience"
```

---

## Self-Review

- Spec coverage: interest imagery is separated from code UI; imagery is used in openers and mid-page chapters; server status is implemented end-to-end; all four pages gain full-screen opener/drawer behavior; cards are limited to the World Deck and LIVE NODE.
- Placeholder scan: no task uses TBD, TODO, or an undefined “similar to” instruction.
- Type consistency: `telemetry.Provider.Snapshot(context.Context) (telemetry.Snapshot, error)` is the only backend boundary; JSON keys exactly match the frontend consumer.
- Safety: public telemetry exposes aggregate metrics only; images are local, licensed, credited, and transformed transparently.
