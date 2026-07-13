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
	source := readIdentitySource(t, "templates/about.tmpl")
	requireIdentityStrings(t, source, "interactive influences",
		`class="influence-preview`,
		`data-influence`,
		`pointerenter`,
		`focusin`,
		`aria-selected`,
	)
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
