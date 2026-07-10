# Personal documentary blog prototype

## Goal

Create a standalone, responsive HTML prototype that makes the blog feel like a
personal documentary rather than a sports campaign. It will retain a light,
calm reading surface while integrating the owner's interests in Max Verstappen,
Stephen Curry, NiKo, and m0NESY through cinematic event stills and restrained
motion.

## Chosen direction

The selected direction is **Private Documentary / Apex & Clutch**. A visitor
never sees all interests competing in one hero. Instead, one scene appears at a
time as an atmospheric memory:

1. The home hero opens with a wet-track F1 still: Verstappen's RB car is small
   at the lower right, with the left side reserved for the blog statement.
2. The featured-essay transition changes to Curry releasing a three-pointer;
   the ball arc becomes the section's progress line.
3. The archive or off-clock transition uses NiKo at a CS stage; a soft crosshair
   aligns briefly with the active post, then fades.

The visible labels `33 / MAX`, `30 / CURRY`, `RIFLER / NIKO`, and `M0NESY`
make the references immediate. There will be no team, manufacturer, sponsor,
or tournament logos.

## Prototype scope

Add a self-contained `prototype-dev/apex-clutch.html` and an adjacent
`prototype-dev/apex-clutch-assets/` directory for the three generated scene
images. Existing prototype files and production templates remain unchanged.

The prototype contains:

- a compact navigation bar with an `off-clock` entry;
- a full-height F1 hero with headline-safe negative space;
- a featured essay panel with Curry imagery and an arc-shaped progress detail;
- a recent-post list whose hover treatment has a short CS crosshair response;
- an `off-clock` index that names the four inspirations plainly;
- a footer plus reduced-motion and mobile states.

## Visual system

- **Paper** `#F3F6FA`: reading-first background.
- **Night cobalt** `#193B78`: depth, navigation, and photographic grade.
- **Mist blue** `#B8CEE2`: quiet transitions and image fades.
- **Race ember** `#E86744`: tail-light and active-state accent.
- **Court gold** `#D9A23B`: Curry cue and small highlights.
- **Radar mint** `#49BFAE`: CS interaction cue.

Typography pairs a restrained Chinese serif display face with a neutral sans
body and a compact monospace utility face. Large imagery has no baked-in copy;
all readable content stays in HTML.

## Motion rules

- The hero car drifts no more than 12 pixels and its rain-light glow pulses
  slowly; there is no autoplay video.
- Scene changes use a 650–800 ms masked fade, never a hard cut or carousel.
- Curry's arc draws once when the featured panel enters the viewport.
- The CS crosshair appears only on the currently hovered or focused post row.
- `prefers-reduced-motion` disables every movement and leaves all scenes visible
  as still, high-contrast imagery.

## Interaction and responsiveness

The page has no runtime data dependency. Navigation and post interactions are
local anchors or prototype links. The desktop composition reserves the left
60% of the hero for copy; on mobile imagery moves beneath the introductory
copy, avoiding cropped faces, cars, or UI controls. All focusable controls have
visible keyboard focus.

## Assets and constraints

Three separate, wide, logo-free, editorial photographic stills are used rather
than a composite image: F1, basketball, and esports. The preview image made
during brainstorming is reference-only; final assets are generated individually
to preserve the one-scene-at-a-time rule. If the images fail to load, gradient
and color fallback layers keep the layout legible.

## Validation

Before handoff, open the standalone file locally and verify desktop and mobile
layouts, scene contrast behind text, keyboard focus, reduced-motion behavior,
and that no asset uses embedded text or logos. The prototype is deliberately
isolated, so production template checks are not required for this phase.
