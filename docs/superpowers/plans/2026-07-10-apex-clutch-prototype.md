# Apex & Clutch Prototype Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a standalone responsive HTML prototype that combines a calm personal blog with one cinematic F1, basketball, or CS scene at a time.

**Architecture:** `prototype-dev/apex-clutch.html` is self-contained: semantic HTML, inline CSS, and a small inline script own the page and its motion states. `prototype-dev/apex-clutch-assets/` holds three logo-free generated stills; a Node test checks the prototype's structural and accessibility contract without adding dependencies or touching production templates.

**Tech Stack:** HTML5, CSS custom properties and keyframes, vanilla JavaScript (`IntersectionObserver`), Node.js built-in test runner, generated PNG assets.

## Global Constraints

- Keep existing `prototype-dev/` pages and all `web/` production templates unchanged.
- Use exactly one prominent photographic scene per section: F1 hero, Curry feature, NiKo archive.
- Do not use team, manufacturer, sponsor, tournament, or broadcast logos; do not embed text in imagery.
- Respect `prefers-reduced-motion`; keep every scene still and all copy legible when it is enabled.
- Retain headline-safe empty space in the left 60% of the desktop hero and move imagery below copy on mobile.

---

### Task 1: Add the static prototype contract test

**Files:**
- Create: `prototype-dev/apex-clutch.test.mjs`
- Create: `prototype-dev/apex-clutch-assets/.gitkeep`
- Create: `prototype-dev/apex-clutch.html`

**Interfaces:**
- Consumes: `prototype-dev/apex-clutch.html` and the three PNG paths defined below.
- Produces: `node --test prototype-dev/apex-clutch.test.mjs` returning one passing test.

- [ ] **Step 1: Write the failing test**

```js
// prototype-dev/apex-clutch.test.mjs
import assert from 'node:assert/strict';
import { readFileSync, existsSync } from 'node:fs';
import test from 'node:test';

const htmlPath = new URL('./apex-clutch.html', import.meta.url);
const assets = [
  new URL('./apex-clutch-assets/f1-verstappen.png', import.meta.url),
  new URL('./apex-clutch-assets/basketball-curry.png', import.meta.url),
  new URL('./apex-clutch-assets/cs-niko.png', import.meta.url),
];

test('apex-clutch prototype has its required scenes and motion fallback', () => {
  const html = readFileSync(htmlPath, 'utf8');
  for (const asset of assets) assert.equal(existsSync(asset), true);
  assert.match(html, /<main[\s>]/);
  assert.match(html, /f1-verstappen\.png/);
  assert.match(html, /basketball-curry\.png/);
  assert.match(html, /cs-niko\.png/);
  assert.match(html, /@media\s*\(prefers-reduced-motion:\s*reduce\)/);
  assert.match(html, /aria-label=/);
});
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `node --test prototype-dev/apex-clutch.test.mjs`

Expected: FAIL with `ENOENT` because `apex-clutch.html` and the assets do not exist yet.

- [ ] **Step 3: Create the asset directory placeholder**

```text
prototype-dev/apex-clutch-assets/.gitkeep
```

- [ ] **Step 4: Commit the test scaffold**

```bash
git add prototype-dev/apex-clutch.test.mjs prototype-dev/apex-clutch-assets/.gitkeep
git commit -m "test: add apex prototype contract"
```

### Task 2: Add the three cinematic stills

**Files:**
- Create: `prototype-dev/apex-clutch-assets/f1-verstappen.png`
- Create: `prototype-dev/apex-clutch-assets/basketball-curry.png`
- Create: `prototype-dev/apex-clutch-assets/cs-niko.png`

**Interfaces:**
- Consumes: the composition and no-logo constraints in `docs/superpowers/specs/2026-07-10-personal-documentary-prototype-design.md`.
- Produces: three readable PNG files referenced by the HTML and contract test.

- [ ] **Step 1: Keep the approved F1 still as the hero asset**

Copy the approved image to `prototype-dev/apex-clutch-assets/f1-verstappen.png`. It must remain a wide, misty wet-track image with the car small at lower right and no embedded text or marks.

- [ ] **Step 2: Generate the basketball still**

Use the built-in image generator with this prompt:

```text
Use case: photorealistic-natural. Asset type: wide website feature background.
Stephen Curry in a blue-and-gold basketball jersey just after a three-point release,
seen from a distant side angle in a softly lit arena. Keep Curry in the far right
third with a clean, pale blue court and atmosphere across the left for HTML copy.
Premium quiet sports editorial photography, filmic 35mm texture, restrained gold
highlight, no visible logos, no text, no watermark, no scoreboard, no crowd detail.
```

Save the selected image as `prototype-dev/apex-clutch-assets/basketball-curry.png`.

- [ ] **Step 3: Generate the CS still**

Use the built-in image generator with this prompt:

```text
Use case: photorealistic-natural. Asset type: wide website archive background.
Nikola "NiKo" Kovač, professional Counter-Strike rifler, focused in side profile at
a tournament desk with headset, hands at keyboard and mouse. Place him in the far
right third; use cool monitor glow, soft blue crowd bokeh, and calm negative space
on the left. Premium esports editorial photography, restrained and contemplative,
no team logos, no text, no watermark, no game HUD, no scoreboard.
```

Save the selected image as `prototype-dev/apex-clutch-assets/cs-niko.png`.

- [ ] **Step 4: Verify the asset contract**

Run: `node --test prototype-dev/apex-clutch.test.mjs`

Expected: FAIL only because `apex-clutch.html` still does not exist.

- [ ] **Step 5: Commit the scene assets**

```bash
git add prototype-dev/apex-clutch-assets
git commit -m "feat: add apex prototype scenes"
```

### Task 3: Build the standalone page

**Files:**
- Create: `prototype-dev/apex-clutch.html`
- Modify: `prototype-dev/apex-clutch.test.mjs`

**Interfaces:**
- Consumes: the three fixed asset paths created in Task 2.
- Produces: a semantic one-page prototype with `data-scene` sections, keyboard-visible post links, and an `is-visible` reveal state.

- [ ] **Step 1: Write the remaining failing assertions**

Append these checks inside the existing test callback:

```js
assert.match(html, /data-scene="f1"/);
assert.match(html, /data-scene="basketball"/);
assert.match(html, /data-scene="cs"/);
assert.match(html, /class="post-link"/);
assert.match(html, /IntersectionObserver/);
```

- [ ] **Step 2: Run the test to verify the new assertions fail**

Run: `node --test prototype-dev/apex-clutch.test.mjs`

Expected: FAIL with the first missing `data-scene` assertion.

- [ ] **Step 3: Implement the page structure and styles**

Create a complete HTML document with these bounded units:

```html
<main>
  <section class="hero scene scene-f1 is-visible" data-scene="f1">...</section>
  <section class="feature scene scene-basketball" data-scene="basketball">...</section>
  <section class="archive scene scene-cs" data-scene="cs">...</section>
</main>
```

Use each PNG as a real `<img alt="">` inside its own scene layer. Add an HTML
overlay (`.scene-wash`) instead of baking text into images. Give the hero an
`h1`, navigation links an `aria-label`, posts the `post-link` class, and the
featured basketball section an inline SVG arc with `aria-hidden="true"`.

Use CSS variables for the six colors defined in the spec. Use `clip-path`,
opacity, and transform only for the scene reveals; scope every animation under
`.motion-ok`. Include a `@media (prefers-reduced-motion: reduce)` block that
removes transition and animation properties and resets hidden scenes to visible.

- [ ] **Step 4: Implement scene activation**

Add this script before `</body>`:

```js
const scenes = document.querySelectorAll('.scene');
if (!matchMedia('(prefers-reduced-motion: reduce)').matches) {
  document.documentElement.classList.add('motion-ok');
  const observer = new IntersectionObserver((entries) => {
    for (const entry of entries) {
      if (entry.isIntersecting) entry.target.classList.add('is-visible');
    }
  }, { threshold: 0.25 });
  scenes.forEach((scene) => observer.observe(scene));
}
```

- [ ] **Step 5: Run the contract test to verify it passes**

Run: `node --test prototype-dev/apex-clutch.test.mjs`

Expected: PASS with one passing subtest.

- [ ] **Step 6: Commit the complete prototype**

```bash
git add prototype-dev/apex-clutch.html prototype-dev/apex-clutch.test.mjs
git commit -m "feat: build apex clutch prototype"
```

### Task 4: Perform visual and accessibility validation

**Files:**
- Modify only if validation exposes a concrete defect: `prototype-dev/apex-clutch.html`

**Interfaces:**
- Consumes: the page and test contract from Tasks 1–3.
- Produces: verified desktop, mobile, keyboard, and reduced-motion behavior.

- [ ] **Step 1: Run the static contract test**

Run: `node --test prototype-dev/apex-clutch.test.mjs`

Expected: PASS with one passing subtest and no warnings.

- [ ] **Step 2: Serve the prototype locally**

Run: `python -m http.server 4173 --directory prototype-dev`

Expected: `Serving HTTP on 0.0.0.0 port 4173`.

- [ ] **Step 3: Inspect the desktop layout**

Open `http://localhost:4173/apex-clutch.html`. Confirm the F1 car remains in the
right 40%, each following scene is singular rather than a collage, and text has
contrast against its wash layer.

- [ ] **Step 4: Inspect mobile and motion fallbacks**

At 390 px width, confirm each image appears beneath its section copy without a
cropped face or car. Emulate reduced motion and confirm all three scenes are
visible with no movement. Tab through the navigation and post links; every focus
target must display an outline.

- [ ] **Step 5: Commit only a validation-driven correction**

```bash
git add prototype-dev/apex-clutch.html
git commit -m "fix: refine apex prototype layout"
```

Run this step only if Steps 3 or 4 identify and correct a concrete defect.
