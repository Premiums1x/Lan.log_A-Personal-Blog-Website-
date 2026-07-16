# Lancer Blog — Agent Design Charter

## Product identity

This repository is not a generic developer-blog template with a personal name attached. It is Lancer's private-facing public notebook: a place built primarily for Lancer to record exploration, setbacks, recovery, technical growth, and the belief that a better tomorrow is earned by continuing forward.

The site must remain identifiable as Lancer's even if the visible name and avatar are removed.

## The person behind the site

- Public identity: **Lancer** only. Do not expose a legal name, school, employer, precise location, personal photograph, or other private identifiers unless the user explicitly authorizes it.
- Primary audience: Lancer himself. External readers are welcome, but the experience must not become a recruiter-first portfolio or a content-marketing site.
- Core temperament: exploratory, energetic, life-loving, resilient, competitive, and hopeful without becoming motivational-poster copy.
- Competitive ideal: break through a defense like an entry fragger, then become calm and dependable in a clutch. The interface should express both controlled aggression and composure under pressure.
- Sports vocabulary: F1, NBA, football, and Counter-Strike. Admired figures include Verstappen, Leclerc, Curry, Messi, NiKo, and TheShy.
- Music vocabulary: millennium-era Chinese pop, Jay Chou, G.E.M., contemporary phonk, and EDM. Music represents the imagined peak moment after winning a clutch, not background decoration.

## Design thesis: Break the line, hold the clutch

The visual system combines four worlds:

1. **Racing line** — velocity, telemetry, precision, and progress through a lap.
2. **Entry route** — the courage to tear open a defense and create space.
3. **Clutch room** — reduced noise, exact information, and trust under pressure.
4. **Epic poster** — a cinematic sense of scale reserved for the hero and chapter openings.

The signature device is the **Lancer Line**: one continuous trajectory that can read as a racing line, a tactical route, an attacking run, or a pulse of forward motion. Use it as a structural guide, transition, focus state, or restrained motion motif. Do not scatter unrelated glows and animations across the page.

## Visual direction

- Desired tension: young and experimental + cold and precise + fast and attacking.
- Default surface: light, high-clarity editorial space rather than a permanently dark terminal.
- Cinematic contrast is allowed in one major place per page, especially the home hero or a chapter opener.
- Suggested palette roles:
  - `Grid Black #0B0D12` — cinematic panels and high-focus moments.
  - `Paper White #F5F4EF` — primary site canvas and reading calm.
  - `Clutch White #FFFFFF` — article paper and elevated content.
  - `Attack Orange #FF4D2E` — breakthrough, primary action, and the Lancer Line.
  - `Telemetry Cyan #67D8FF` — data, exploration, and secondary signals.
  - `Trophy Gold #D4A94F` — rare milestones only; never a general accent.
- Typography should pair a condensed, poster-like display face with a highly readable Chinese body face and a mono utility face. Do not use one neutral sans family for every role.
- Cards may use drawers, sliding reveals, clipped corners, stacked depth, or track-like transitions when they encode content. Avoid generic floating SaaS cards.

## Existing concepts to preserve through reinterpretation

- Terminal and code language: retain as a utility dialect, not the entire visual identity.
- Commit metadata: retain as provenance and progress markers.
- Article body: keep a clearly independent, calm reading module with strong boundaries.
- GitHub contribution heatmap: retain as a live training record; if unavailable, leave the area empty rather than restoring a biography fallback.
- Motion: purposeful, state-based, and resilient during fast scrolling. Preserve `prefers-reduced-motion` support.

## Reference translation

- `landonorris.com`: learn from its character-led storytelling, On Track / Off Track separation, bold image/card composition, drawers, layered transitions, and premium pacing. Do not copy McLaren/Lando branding, lime color, signatures, photos, or layout verbatim.
- `cyhkbl.qzz.io`: learn from the willingness to place personal interests directly on the homepage. Lancer's interests should be visible rather than hidden inside an About paragraph.
- `sibuchen.xyz`: learn from its sense of finish, spatial confidence, and authored design. Do not reproduce its visual assets or component arrangement.
- Athlete, player, team, game, anime, and music imagery requires user-supplied or properly licensed assets. Until then, use original code-native graphics, typography, telemetry, abstract tactics, and user-authored copy.

## Information architecture

- Home is a personal cockpit and current chapter, not merely a hero followed by recent posts.
- Posts are field notes / rounds / laps from an ongoing life, while remaining ordinary accessible articles in URLs and semantics.
- Archive is a season record showing change over time.
- About should reveal values, influences, and operating principles without exposing protected identity details.
- Shelf can hold influences across technology, sport, games, and music rather than only educational resources.

## Copy voice

- Chinese-first, with English used for concise telemetry, commands, labels, and established sports/game terms.
- Direct, specific, self-aware, and quietly hopeful.
- Competitive language is welcome, but avoid hollow slogans, fake achievements, exaggerated confidence, and fan-wiki biography.
- Never write as if Lancer is already a professional athlete, esports player, celebrity, or expert he has not claimed to be.
- Prefer lived-state copy such as "还在进攻", "这一回合没结束", and "把今天推进一点" over generic lines such as "热爱生活，追逐梦想".

## Interaction rules

- Spend boldness on one orchestrated moment per page.
- Motion should reveal hierarchy or direction: drawer opening, route progression, card handoff, or chapter change.
- No constant high-frequency border effects, cursor chasers, scroll-jacking, autoplay audio, fake loading screens, or motion that competes with reading.
- Keep keyboard focus visible, touch targets usable, layouts responsive, and all core information accessible without animation.
- Fast scrolling must not produce flicker, runaway transforms, or rapid effect retriggering.

## Engineering constraints

- Public UI source of truth: `web/templates/*.tmpl` and `web/static/style.css`.
- Homepage and secondary-page copy is database-backed through `settings`; template edits alone do not overwrite existing deployed settings.
- Admin UI source: `web/admin/src`.
- Preserve explicit blank excerpts. Never generate or display body-text fallbacks as summaries.
- Add or update contract tests before changing production templates, interactions, or content behavior.
- Required verification for public UI changes:
  - `go test ./... -count=1`
  - `go run ./cmd/tplcheck`
  - Visual QA at desktop and mobile widths
  - `npm run build` in `web/admin` when admin sources change
- Production deployment embeds `web/admin-dist` and migrations in the Go binary; run the appropriate deploy build before upload.

## Acceptance test

Before considering a design complete, ask:

1. If the name and avatar disappeared, would this still feel specifically like Lancer?
2. Does it express both attacking energy and clutch calm?
3. Does it reveal genuine interests without becoming a collage of fandom logos?
4. Is the article-reading experience still quieter than the surrounding site?
5. Did every visual flourish earn its place through identity, structure, or usability?

If any answer is no, revise before shipping.
