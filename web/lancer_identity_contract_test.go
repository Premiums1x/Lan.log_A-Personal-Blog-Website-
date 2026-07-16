package web

import (
	"os"
	"strings"
	"testing"
)

func readIdentitySource(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

func requireIdentityStrings(t *testing.T, source, label string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(source, want) {
			t.Errorf("%s missing %q", label, want)
		}
	}
}

func TestPublicShellLoadsLancerIdentityLayer(t *testing.T) {
	source := readIdentitySource(t, "templates/layout.tmpl")
	requireIdentityStrings(t, source, "public shell",
		`/static/lancer.css`,
		`class="lancer-site`,
		`class="lancer-mark">L/`,
		`FIELD NOTES`,
	)
}

func TestHomeUsesLancerCockpitAndTrajectory(t *testing.T) {
	source := readIdentitySource(t, "templates/index.tmpl")
	requireIdentityStrings(t, source, "home cockpit",
		`class="lancer-hero`,
		`class="lancer-line`,
		`EXPLORE`,
		`BREAK`,
		`CLUTCH`,
		`NEXT`,
		`CURRENT PLAY`,
		`FIELD NOTES`,
		`INFLUENCE DECK`,
	)
}

func TestIdentityCSSDefinesRequiredTokensAndReducedMotion(t *testing.T) {
	source := readIdentitySource(t, "static/lancer.css")
	requireIdentityStrings(t, source, "identity CSS",
		`--grid-black: #0B0D12`,
		`--paper-white: #F5F4EF`,
		`--attack-orange: #FF4D2E`,
		`--telemetry-cyan: #67D8FF`,
		`@media (prefers-reduced-motion: reduce)`,
	)
}

func TestSecondaryPagesUseLancerChapters(t *testing.T) {
	checks := map[string][]string{
		"templates/about.tmpl":    {`PLAYER PROFILE`, `ON TRACK`, `OFF TRACK`, `CURRENT ROUND`},
		"templates/archive.tmpl":  {`SEASON LOG`, `RACE CONTROL`, `class="season-line`},
		"templates/shelf.tmpl":    {`LOADOUT`, `INFLUENCE DECK`, `class="loadout-card`},
		"templates/notfound.tmpl": {`OFF ROUTE`, `class="off-route-panel`},
	}
	for file, wants := range checks {
		source := readIdentitySource(t, file)
		requireIdentityStrings(t, source, file, wants...)
	}
}

func TestProgrammerIdentityRunsThroughPublicPages(t *testing.T) {
	checks := map[string][]string{
		"templates/index.tmpl":   {`class="dev-console`, `entry.route.ts`, `lancer@field-notes`, `class="terminal-stream`},
		"templates/about.tmpl":   {`profile.manifest`, `stack.current`},
		"templates/archive.tmpl": {`git log --all`, `class="archive-terminal`},
		"templates/shelf.tmpl":   {`ls ./loadout`, `class="loadout-terminal`},
	}
	for file, wants := range checks {
		source := readIdentitySource(t, file)
		requireIdentityStrings(t, source, file, wants...)
	}
}

func TestPrimaryOpenersUseOpenCinematicStages(t *testing.T) {
	for _, file := range []string{"templates/index.tmpl", "templates/about.tmpl", "templates/archive.tmpl", "templates/shelf.tmpl"} {
		source := readIdentitySource(t, file)
		requireIdentityStrings(t, source, file, `cinematic-stage`)
	}
	css := readIdentitySource(t, "static/lancer.css")
	requireIdentityStrings(t, css, "open cinematic stage", `.cinematic-stage`, `border-radius: 0`)
}

func TestPrimaryPagesUseViewportOpenersAndDrawerShells(t *testing.T) {
	for _, file := range []string{"templates/index.tmpl", "templates/about.tmpl", "templates/archive.tmpl", "templates/shelf.tmpl"} {
		source := readIdentitySource(t, file)
		requireIdentityStrings(t, source, file, `page-opening`, `class="drawer-cue`, `class="page-drawer`, `data-drawer-target`, `class="drawer-handle"`, `data-drawer-handle`)
	}
	css := readIdentitySource(t, "static/lancer.css")
	requireIdentityStrings(t, css, "drawer CSS", `min-height: calc(100svh`, `@media (min-width: 1081px) and (min-height: 720px)`, `height: calc(100svh - var(--lancer-nav-height, 76px))`, `position: sticky`, `top: var(--lancer-nav-height`, `.drawer-handle`, `touch-action: none`, `cursor: grab`, `@media (prefers-reduced-motion: reduce)`)
}

func TestDrawerCueUsesSharedProgressiveBehavior(t *testing.T) {
	layout := readIdentitySource(t, "templates/layout.tmpl")
	requireIdentityStrings(t, layout, "public shell", `/static/lancer.js`)

	source := readIdentitySource(t, "static/lancer.js")
	requireIdentityStrings(t, source, "drawer interactions", `[data-drawer-handle]`, `addEventListener('pointerdown'`, `addEventListener('pointermove'`, `setPointerCapture`, `requestAnimationFrame`, `window.scrollTo`, `startScrollY + dragDistance`, `scrollBehavior = 'auto'`)
	if strings.Contains(source, `classList.add('is-used')`) {
		t.Fatal("drawer cue must remain fully visible after it has been used")
	}
	if strings.Contains(source, `getBoundingClientRect()`) || strings.Contains(source, `--drawer-progress`) || strings.Contains(source, `dragDistance * 1.65`) {
		t.Fatal("drawer must preserve native sticky scrolling outside direct handle dragging")
	}

	css := readIdentitySource(t, "static/lancer.css")
	requireIdentityStrings(t, css, "drawer motion", `font-size: 1.75rem`, `animation: drawer-cue-float 3.4s cubic-bezier(.45, 0, .2, 1) infinite`, `@keyframes drawer-cue-float`)
	if strings.Contains(css, `drawer-cue-step`) || strings.Contains(css, `steps(2, end)`) {
		t.Fatal("drawer cue must use slow continuous motion rather than stepped jitter")
	}
}

func TestCinematicStagesDoNotReflowAbsoluteAtmosphereLayers(t *testing.T) {
	css := readIdentitySource(t, "static/lancer.css")
	if strings.Contains(css, `.cinematic-stage > *:not(.cinematic-image) { position: relative;`) {
		t.Fatal("cinematic stage must not turn absolute atmosphere layers into grid items")
	}
}
func TestCinematicBackgroundAssetsAreWired(t *testing.T) {
	css := readIdentitySource(t, "static/lancer.css")
	for _, asset := range []string{"lancer-dev-cockpit.webp", "lancer-code-runway.webp"} {
		requireIdentityStrings(t, css, "cinematic asset CSS", asset)
		if _, err := os.Stat("static/media/" + asset); err != nil {
			t.Errorf("cinematic asset %s missing: %v", asset, err)
		}
	}
}

func TestAboutInfluencesSwitchOnPointerAndKeyboardFocus(t *testing.T) {
	templateSource := readIdentitySource(t, "templates/about.tmpl")
	requireIdentityStrings(t, templateSource, "interactive influence markup",
		`class="influence-preview`,
		`data-influence`,
		`data-image`,
		`role="group"`,
		`aria-pressed`,
	)
	if strings.Contains(templateSource, `role="listbox"`) || strings.Contains(templateSource, `aria-selected`) {
		t.Fatal("influence buttons must use ordinary group and pressed-state semantics")
	}

	scriptSource := readIdentitySource(t, "static/lancer.js")
	requireIdentityStrings(t, scriptSource, "interactive influence behavior",
		`.interactive-influences`,
		`pointerenter`,
		`focusin`,
		`click`,
		`getAttribute('aria-pressed')`,
		`setAttribute('aria-pressed'`,
		`layers[0].style.setProperty('--influence-image'`,
		`layers[nextLayer].style.setProperty('--influence-image'`,
	)
	if strings.Contains(scriptSource, `aria-selected`) {
		t.Fatal("influence behavior must update aria-pressed rather than aria-selected")
	}
	if strings.Contains(scriptSource, `root.style.setProperty('--influence-image'`) {
		t.Fatal("influence image state must be isolated per preview layer")
	}
}

func TestShelfDisclosuresWorkWithKeyboardAndWithoutLinks(t *testing.T) {
	templateSource := readIdentitySource(t, "templates/shelf.tmpl")
	requireIdentityStrings(t, templateSource, "shelf disclosure markup", `class="loadout-card loadout-card-static"`)

	cssSource := readIdentitySource(t, "static/lancer.css")
	requireIdentityStrings(t, cssSource, "shelf keyboard disclosure CSS",
		`.loadout-card:focus-visible`,
		`.loadout-card:focus-within::before`,
		`.loadout-card:focus-within .loadout-drawer`,
		`.loadout-card-static .loadout-drawer`,
	)
}

func TestArticleListsCollapseAfterSixAndKeepTruePublishedCount(t *testing.T) {
	index := readIdentitySource(t, "templates/index.tmpl")
	requireIdentityStrings(t, index, "home article expansion",
		`data-expand-list data-expand-initial="5"`,
		`data-expand-item`,
		`data-expand-trigger hidden`,
		`aria-expanded="false"`,
		`展开余下文章`,
		`{{.PostCount}}`,
	)

	archive := readIdentitySource(t, "templates/archive.tmpl")
	requireIdentityStrings(t, archive, "archive season expansion",
		`class="season-year" data-expand-list data-expand-initial="6"`,
		`data-expand-item`,
		`data-expand-trigger hidden`,
		`展开余下文章`,
	)

	script := readIdentitySource(t, "static/lancer.js")
	requireIdentityStrings(t, script, "progressive article expansion",
		`[data-expand-list]`,
		`[data-expand-item]`,
		`[data-expand-trigger]`,
		`items.slice(initialCount)`,
		`trigger.hidden = false`,
		`trigger.setAttribute('aria-expanded', 'true')`,
	)

	css := readIdentitySource(t, "static/lancer.css")
	requireIdentityStrings(t, css, "article expansion styling",
		`.article-expand`,
		`.field-card[hidden]`,
		`.season-note[hidden]`,
	)

	repoSource := readIdentitySource(t, "../internal/repo/repo.go")
	requireIdentityStrings(t, repoSource, "published count repository",
		`func CountPublished`,
		`SELECT COUNT(*) FROM posts WHERE status='published'`,
		`func ListPublishedAll`,
	)

	handlerSource := readIdentitySource(t, "../internal/handler/public.go")
	requireIdentityStrings(t, handlerSource, "home true published count",
		`repo.ListPublishedAll`,
		`repo.CountPublished`,
		`PostCount: postCount`,
	)
	if strings.Contains(handlerSource, `PostCount: len(posts)`) {
		t.Fatal("home must not report the length of a truncated result as the published total")
	}
}
func TestHomeInfluenceDeckRevealsUserImagesWithTheExistingDrawerMotion(t *testing.T) {
	templateSource := readIdentitySource(t, "templates/index.tmpl")
	requireIdentityStrings(t, templateSource, "home influence deck accessibility", `class="influence-card reveal influence-{{$i}}" tabindex="0"`)

	css := readIdentitySource(t, "static/lancer.css")
	requireIdentityStrings(t, css, "home influence image reveal",
		`--influence-image: url('/static/media/interests/f1-deck-user.jpg')`,
		`--influence-image: url('/static/media/interests/nba-deck-user.jpg')`,
		`--influence-position: 50% 75%`,
		`--influence-image: url('/static/media/interests/cs2-deck-user.jpg')`,
		`--influence-image: url('/static/media/interests/jay-deck-user.jpg')`,
		`.influence-card::after`,
		`background-image: linear-gradient`,
		`transform: translateY(calc(100% - 5px)) scale(1.025)`,
		`.influence-card:hover::after`,
		`.influence-card:focus-visible::after`,
		`opacity: .78`,
	)
}
func TestHomeUsesSeparateMobileOpeningFocalPoint(t *testing.T) {
	templateSource := readIdentitySource(t, "templates/index.tmpl")
	requireIdentityStrings(t, templateSource, "home opener crop",
		`--opening-position:center 22%`,
		`--opening-position-mobile:54% 20%`,
	)

	cssSource := readIdentitySource(t, "static/lancer.css")
	requireIdentityStrings(t, cssSource, "mobile opener crop CSS",
		`background-position: var(--opening-position-mobile, var(--opening-position))`,
	)
}

func TestInterestWorldAppearsBeyondEveryHero(t *testing.T) {
	checks := map[string][]string{
		"templates/index.tmpl":   {`class="world-deck`, `data-world="counter-strike"`, `class="chapter-image-band`, `niko-user.jpg`},
		"templates/about.tmpl":   {`--influence-image`, `niko-user.jpg`, `verstappen-user.jpg`},
		"templates/archive.tmpl": {`class="chapter-image-band`, `leclerc-user.jpg`},
		"templates/shelf.tmpl":   {`class="world-card`, `messi-user.jpg`},
		"templates/layout.tmpl":  {`MEDIA / USER PROVIDED`},
	}
	for file, wants := range checks {
		requireIdentityStrings(t, readIdentitySource(t, file), file, wants...)
	}
}
func TestHomeWorldCardsResetToEqualWidthAndUseRequestedSubjects(t *testing.T) {
	templateSource := readIdentitySource(t, "templates/index.tmpl")
	requireIdentityStrings(t, templateSource, "home world cards", `>Counter-Strike</h3>`, `>F1</h3>`, `>Basketball</h3>`, `curry-new-user.jpg`, `--world-position:50% 64%`)
	if strings.Contains(templateSource, `world-card is-active`) || strings.Contains(templateSource, `data-world="football"`) {
		t.Fatal("home world cards must start equal and must not retain the football card")
	}

	cssSource := readIdentitySource(t, "static/lancer.css")
	if strings.Contains(cssSource, `.world-card.is-active { flex-grow`) {
		t.Fatal("world deck must not keep one card expanded after pointer exit")
	}
}
func TestHomeCockpitAndWorldCardCopyRespectImageSpace(t *testing.T) {
	css := readIdentitySource(t, "static/lancer.css")
	requireIdentityStrings(t, css, "home image-safe layout",
		`width: min(100%, 760px)`,
		`margin-right: clamp(-56px, -3vw, -24px)`,
		`.world-card[data-world="counter-strike"] h3`,
		`.world-card[data-world="f1"] h3`,
		`letter-spacing: .08em`,
		`white-space: nowrap`,
		`max-width: none`,
	)
}
func TestHomeNiKoImagesUsePosterScaleAndImageAwareChapterCopy(t *testing.T) {
	templateSource := readIdentitySource(t, "templates/index.tmpl")
	requireIdentityStrings(t, templateSource, "home world deck trophy media", `niko-trophy-user.jpg`)

	css := readIdentitySource(t, "static/lancer.css")
	requireIdentityStrings(t, css, "home NiKo composition",
		`background-size: auto 125%`,
		`.world-deck-heading`,
		`background-size: auto 92%`,
		`.world-deck-heading .chapter-band-copy`,
		`margin-left: auto`,
		`text-align: right`,
		`.dev-stage .cinematic-image::before`,
	)
}
func TestAboutHeroAndInfluencesKeepPeopleVisible(t *testing.T) {
	templateSource := readIdentitySource(t, "templates/about.tmpl")
	requireIdentityStrings(t, templateSource, "about requested media", `data-influence="sport"`, `curry-new-user.jpg`, `data-position="50% 58%"`, `data-influence="music"`, `data-image="/static/media/interests/gem-user.jpg" data-position="50% 42%" data-size="cover"`)

	css := readIdentitySource(t, "static/lancer.css")
	requireIdentityStrings(t, css, "about image-safe layout", `.about-stage .cinematic-image`, `background-size: auto 122%`, `width: min(100%, 620px)`, `background-repeat: no-repeat`)

	js := readIdentitySource(t, "static/lancer.js")
	requireIdentityStrings(t, js, "influence crop runtime", `dataset.size || 'cover'`, `backgroundSize`)
}
func TestArchiveAndShelfOpenersPreservePortraitSubjects(t *testing.T) {
	css := readIdentitySource(t, "static/lancer.css")
	requireIdentityStrings(t, css, "secondary opener image safety",
		`.archive-stage .cinematic-image`,
		`background-size: auto 125%`,
		`background: var(--opening-image)`,
		`mask-image: linear-gradient`,
		`z-index: 0`,
		`z-index: 2`,
		`.shelf-stage .cinematic-image`,
		`background-size: auto 100%`,
		`.archive-stage .archive-terminal`,
		`.shelf-stage .loadout-terminal`,
		`background-position: 18% center`,
		`width: min(100%, 560px)`,
		`width: min(100%, 620px)`,
	)
}
func TestArchiveChapterKeepsCopyClearOfLeclerc(t *testing.T) {
	templateSource := readIdentitySource(t, "templates/archive.tmpl")
	requireIdentityStrings(t, templateSource, "archive chapter media layers", `class="archive-chapter-media archive-chapter-backdrop"`, `class="archive-chapter-media archive-chapter-portrait"`)

	css := readIdentitySource(t, "static/lancer.css")
	requireIdentityStrings(t, css, "archive chapter composition", `.archive-chapter-backdrop`, `linear-gradient(180deg, #9AA8B6 0%`, `.archive-chapter-portrait`, `background-size: auto 118%`, `mask-image: linear-gradient`, `background-image: none`)
	if strings.Contains(css, `background-image: var(--chapter-image), var(--chapter-image)`) || strings.Contains(css, `filter: blur(28px)`) {
		t.Fatal("archive chapter must use one crisp portrait over a photo-matched silver field")
	}
}
func TestArticleKeepsCalmPaperInsideLancerIdentity(t *testing.T) {
	source := readIdentitySource(t, "templates/post.tmpl")
	requireIdentityStrings(t, source, "article identity",
		`ROUND NOTES`,
		`class="article-cockpit`,
		`class="article-paper`,
		`{{if .Post.Excerpt}}`,
		`{{if .Heatmap}}`,
	)
	if strings.Contains(source, `article-paper reveal`) {
		t.Fatal("article reading paper must not use the global reveal animation")
	}
}

func TestHomeRendersLiveNodeWithoutLeakingHostIdentity(t *testing.T) {
	source := readIdentitySource(t, "templates/index.tmpl")
	requireIdentityStrings(t, source, "live node",
		`LIVE NODE`,
		`data-system-status`,
		`data-metric="cpu"`,
		`data-metric="memory"`,
		`data-metric="frequency"`,
		`data-metric="uptime"`,
		`data-chart="cpu"`,
		`data-chart="frequency"`,
		`data-chart="memory"`,
		`data-chart="uptime"`,
		`data-chart-line`,
		`data-chart-area`,
	)

	js := readIdentitySource(t, "static/lancer.js")
	requireIdentityStrings(t, js, "telemetry client", `/api/system-status`, `AbortController`, `visibilityState`, `const sampleLimit = 40`, `chartHistory`, `frequencyCeiling`, `Math.max(5000`, `key === 'uptime'`, `height / 2`, `appendChartPoint`, `setAttribute('d'`)
	for _, forbidden := range []string{"hostname", "processes", "mounts"} {
		if strings.Contains(strings.ToLower(source+js), forbidden) {
			t.Fatalf("public telemetry references %s", forbidden)
		}
	}
}

func TestLiveNodePollingIsFailureHonestAndResponsive(t *testing.T) {
	js := readIdentitySource(t, "static/lancer.js")
	requireIdentityStrings(t, js, "telemetry safety",
		`15000`,
		`8000`,
		`setInterval`,
		`clearInterval`,
		`clearTimeout`,
		`const timeout = setTimeout`,
		`clearTimeout(timeout)`,
		`visibilitychange`,
		`textContent`,
		`UNAVAILABLE`,
		`Math.max(0, Math.min(100`,
	)

	css := readIdentitySource(t, "static/lancer.css")
	requireIdentityStrings(t, css, "telemetry responsive CSS",
		`.live-node-card`,
		`.live-node-grid`,
		`.telemetry-metric-head`,
		`.telemetry-chart-line`,
		`.telemetry-chart-area`,
		`grid-template-columns: repeat(2, minmax(0, 1fr))`,
		`min-height: 205px`,
		`@media (max-width: 320px)`,
	)
}

func TestLiveNodeUptimeStaysOnOneLine(t *testing.T) {
	css := readIdentitySource(t, "static/lancer.css")
	requireIdentityStrings(t, css, "uptime metric", `.live-node-grid [data-metric="uptime"]`, `white-space: nowrap`, `font-size: clamp(18px, 1.8vw, 24px)`)
}

func TestArchiveSupportingTypeRemainsReadable(t *testing.T) {
	css := readIdentitySource(t, "static/lancer.css")
	requireIdentityStrings(t, css, "archive supporting typography",
		`.race-control-head`, `font: 500 11px var(--telemetry)`,
		`.control-block > a`, `font-size: 14px`,
		`.control-block h2`, `font: 500 10px var(--telemetry)`,
		`.control-tags a`, `font: 500 11px var(--telemetry)`,
		`.season-note-meta`, `font: 500 14px var(--telemetry)`,
		`.season-note-copy > p`, `font-size: 15px`,
		`.season-note-copy > div span`, `font: 500 13px var(--telemetry)`,
		`.season-note-read`, `font: 500 11px var(--telemetry)`,
		`.season-note-read b`, `font-size: 28px`,
	)
}
func TestArchiveTerminalHeaderAndOrderControlStayUsable(t *testing.T) {
	templateSource := readIdentitySource(t, "templates/archive.tmpl")
	requireIdentityStrings(t, templateSource, "archive order control",
		`class="archive-order-control"`,
		`data-archive-sort`,
		`data-archive-order-label`,
		`data-archive-year`,
		`data-published=`,
		`aria-label="切换文章排序，当前从新到旧"`,
	)

	css := readIdentitySource(t, "static/lancer.css")
	requireIdentityStrings(t, css, "archive terminal header repair",
		`.archive-terminal > .dev-editor-head`,
		`height: 52px`,
		`.archive-terminal > .dev-editor-head > span`,
		`align-self: center`,
		`.archive-terminal > .dev-editor-head > strong`,
		`font: 700 clamp(20px, 2vw, 28px)/1.2 var(--display)`,
		`.archive-order-control`,
		`.archive-order-control:focus-visible`,
	)

	script := readIdentitySource(t, "static/lancer.js")
	requireIdentityStrings(t, script, "archive date sorting",
		`[data-archive-sort]`,
		`[data-archive-year]`,
		`[data-published]`,
		`OLDEST`,
		`NEWEST`,
		`aria-label`,
	)
}

func TestArchiveRaceControlSupportsPartialNav(t *testing.T) {
	templateSource := readIdentitySource(t, "templates/archive.tmpl")
	requireIdentityStrings(t, templateSource, "archive race-control partial nav",
		`href="/archive" data-archive-nav`,
		`href="/section/{{.Slug}}" data-archive-nav`,
		`href="/tags/{{.Slug}}" data-archive-nav`,
	)

	script := readIdentitySource(t, "static/lancer.js")
	requireIdentityStrings(t, script, "archive partial nav client",
		`[data-archive-nav]`,
		`DOMParser`,
		`parseFromString`,
		`history.pushState`,
		`popstate`,
		`initExpandLists`,
	)
}
func TestLancerIdentityMigrationUpdatesExistingSettings(t *testing.T) {
	source := readIdentitySource(t, "migrations/0004_lancer_identity.sql")
	requireIdentityStrings(t, source, "identity migration",
		`UPDATE settings`,
		`'hero'`,
		`'stack'`,
		`'about'`,
		`'now'`,
		`'archive'`,
		`'shelf'`,
		`Lancer`,
		`这一回合没结束`,
	)
}
