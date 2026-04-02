---
name: stream-deck
description: >
  Create slide-deck presentation webs for streams and courses using Hefesto Forge theme with inline SVG diagrams.
  Trigger: When building a presentation, slide deck, course material, stream web, or talk slides.
metadata:
  author: gentleman-programming
  version: "1.1"
---

## When to Use

- Building a slide-deck web presentation for streams, talks, or courses
- Creating inline SVG diagrams for dark-themed presentations
- Setting up a Hefesto Forge themed web UI
- Generating visual diagrams with high contrast on dark backgrounds

---

## Architecture Overview

Single-page HTML presentation with:
- **No frameworks** — vanilla HTML/CSS/JS
- **No build step** — open `index.html` directly
- **No vertical scroll** — `100dvh` viewport, everything fits
- **Inline SVGs** — all diagrams are SVG elements in HTML (no image files)
- **Module system** — slides grouped into modules, displayed in a sidebar rail
- **Vim-mode theming** — lualine-inspired mode badges (Normal, Command, Insert, Visual, Terminal, Replace)

```
project/
├── index.html              # Single HTML file with all slides
└── assets/
    ├── css/styles.css      # Hefesto Forge theme + layout
    └── js/app.js           # Navigation, dots, mode switching
```

---

## Critical Patterns

### Pattern 1: Hefesto Forge Color Palette

ALWAYS use these exact colors. Source: `HefestoOpenCode/themes/hefesto.json`

```css
:root {
  /* Backgrounds - warm forge blacks */
  --bg: #0a0a0f;
  --bg-dark: #0a0a0f;
  --black: #0a0a0f;
  --gray0: #0d0d12;         /* Rail/viewport background */
  --gray1: #12121a;         /* Card backgrounds, surface */
  --gray2: #1a1a22;         /* Inner panels, surface2 */
  --gray3: #22222a;         /* Deeper surface */
  --line: #3D3529;          /* Borders, separators */
  --line-strong: #6B6358;   /* Strong borders, visible dots */
  --selection: #2a221a;     /* Active module highlight - amber tinted */

  /* Text - CONTRAST IS CRITICAL */
  --fg: #F5F0E8;            /* Primary text — warm white, high contrast */
  --subtext1: #A1AABB;      /* Secondary text (paragraphs) — ~5.2:1 ratio */
  --subtext: #8B8175;       /* Tertiary text (eyebrow, hints) — ~4.2:1 ratio */

  /* Accent colors - forge palette */
  --red: #C4453A;
  --green: #4ADE80;
  --yellow: #E8A84C;        /* Golden/info color */
  --purple: #C17F59;        /* Copper/secondary */
  --magenta: #E87040;       /* Warm orange-red */
  --orange: #E8850C;        /* Primary amber */
  --blue: #4ADE80;          /* Reuse green for cool accent */
  --cyan: #DEBA87;          /* Warning/golden */
  --accent: #E8850C;        /* Mode badge, counter, kicker - primary amber */
}
```

### CRITICAL: Contrast Rules

| Use Case | WRONG Color | CORRECT Color | Why |
|----------|-------------|---------------|-----|
| Muted text on dark bg | `#3D3529` | `#8B8175` | 3D has ~1.5:1 ratio — INVISIBLE |
| Secondary text | `#6B6358` | `#A1AABB` | 6B is textMuted, too dark for secondary |
| Yellow/gold | `#FFE066` | `#E8A84C` | FFE is neon, E8A8 matches forge theme |
| Dot borders | `#3D3529` | `#6B6358` | 3D disappears on dark backgrounds |

**Minimum contrast ratio: 4:1 against `#12121a` surface backgrounds.**

### Pattern 2: Slide HTML Structure

Each slide is a two-column grid with text left, diagram right:

```html
<article class="slide" data-index="0" data-module="0" data-tone="orange">
  <div class="slide-content">
    <p class="slide-kicker">01 · MODULE NAME</p>
    <h2>Slide Title</h2>
    <p>Explanation paragraph.</p>
  </div>
  <figure class="slide-figure">
    <!-- INLINE SVG goes here — never <img> tags -->
    <svg viewBox="0 0 520 360" xmlns="http://www.w3.org/2000/svg"
         font-family="Space Grotesk,sans-serif">
      <!-- diagram content -->
    </svg>
  </figure>
</article>
```

**Key attributes:**
- `data-index` — global slide number (0-based)
- `data-module` — which module group (maps to rail sidebar)
- `data-tone` — color accent for that slide (orange, green, red, etc.)

### Pattern 3: Module/Rail System

Modules are groups of related slides. The sidebar shows module titles + dot indicators:

```html
<nav class="rail">
  <div class="rail-module" data-module="0">
    <button class="rail-title" data-first="0">1. Module Name</button>
    <div class="rail-dots" id="dots-0"></div>
  </div>
  <!-- dots are generated dynamically by JS -->
</nav>
```

The `modeMap` in JS maps module indices to vim modes:

```js
const modeMap = {
  0: "normal",   // orange/amber
  1: "command",  // accent/gold
  2: "insert",   // green
  3: "visual",   // magenta
  4: "replace",  // red
  5: "terminal", // cyan
  6: "normal",   // cycles back
  7: "command",
  8: "insert",
  9: "visual"
};
```

### Pattern 4: Inline SVG Design System

**ALL diagrams are inline SVGs, never image files.**

```
viewBox="0 0 520 360"
Background: transparent (parent container provides #0a0a0f via CSS)
Font: Space Grotesk,sans-serif for ALL text
```

**SVG element conventions:**

| Element | Style |
|---------|-------|
| Section headers | `font-size="10" letter-spacing="2" fill="#A1AABB"` uppercase |
| Titles | `font-size="15-18" font-weight="600" fill="#F5F0E8"` |
| Card backgrounds | `fill="#12121a" stroke="#3D3529"` rounded `rx="8-14"` |
| Inner panels | `fill="#1A1A22"` |
| Subtitle/labels | `fill="#A1AABB"` |
| Muted/decorative | `fill="#8B8175"` |
| Borders | `stroke="#3D3529"` |
| Glows | `feGaussianBlur` filters with accent color, opacity 0.2-0.3 |

**Filter ID rule:** prefix ALL filter IDs with `s{slideIndex}-` to avoid conflicts:
```xml
<filter id="s12-glow"> <!-- slide 12 -->
<filter id="s27-shadow"> <!-- slide 27 -->
```

### Pattern 5: Sub-Agent SVG Generation

Generate SVGs using sub-agents (Task tool) for parallelism. Each sub-agent gets:

1. The full design system (viewBox, palette, font, conventions)
2. The specific slide concept and content
3. Instructions to return ONLY raw `<svg>...</svg>` markup
4. The filter ID prefix for that slide

Then use `mcp_edit` to replace `<img>` tags with the returned SVG.

---

## Decision Tree

```
Need a new presentation?
  → Scaffold HTML with topbar, rail, viewport, controls
  → Define modules and slide count
  → Create CSS with full Hefesto Forge palette
  → Create JS with navigation + buildDots()

Adding slides to existing deck?
  → Add <article class="slide"> with correct data-index, data-module, data-tone
  → Generate inline SVG via sub-agent with design system prompt
  → Insert SVG into <figure class="slide-figure">

Fixing contrast issues?
  → Check fill/stroke colors against the contrast table above
  → Use replaceAll to swap colors globally
  → Verify CSS variables match the corrected palette
```

---

## Code Examples

### Full Slide with Inline SVG

```html
<article class="slide" data-index="5" data-module="1" data-tone="yellow">
  <div class="slide-content">
    <p class="slide-kicker">06 · Context Window</p>
    <h2>Ventana limitada, impacto total</h2>
    <p>El contexto es RAM finita. Cada token que entra empuja otro afuera.</p>
  </div>
  <figure class="slide-figure">
    <svg viewBox="0 0 520 360" xmlns="http://www.w3.org/2000/svg"
         font-family="Space Grotesk,sans-serif" fill="none"
         role="img" aria-label="Ventana de contexto limitada">
      <defs>
        <filter id="s5-glow" x="-50%" y="-50%" width="200%" height="200%">
          <feGaussianBlur stdDeviation="4" result="blur"/>
          <feFlood flood-color="#E8A84C" flood-opacity="0.25"/>
          <feComposite in2="blur" operator="in"/>
          <feMerge><feMergeNode/><feMergeNode in="SourceGraphic"/></feMerge>
        </filter>
      </defs>
      <text x="260" y="30" text-anchor="middle" font-size="10"
            letter-spacing="2" fill="#A1AABB">CONTEXTO</text>
      <rect x="60" y="50" width="400" height="260" rx="12"
            fill="#12121a" stroke="#3D3529" stroke-width="1"/>
      <!-- ... diagram content ... -->
    </svg>
  </figure>
</article>
```

### CSS Layout Grid (No Scroll)

```css
.deck-app {
  height: 100dvh;
  max-height: 100dvh;
  display: grid;
  grid-template-rows: auto auto 1fr auto; /* topbar, progress, stage, controls */
  gap: 14px;
}

.stage-layout {
  min-height: 0;
  display: grid;
  grid-template-columns: 220px 1fr; /* rail + viewport */
  overflow: hidden;
}

.slide {
  position: absolute;
  inset: 0;
  display: grid;
  grid-template-columns: minmax(280px, 36%) 1fr; /* text + diagram */
}
```

### Sub-Agent Prompt Template for SVG Generation

```
You are an SVG diagram generator. Generate a SINGLE inline SVG.

## DESIGN SYSTEM (MANDATORY)
- viewBox="0 0 520 360", NO background rect (transparent)
- Font: Space Grotesk,sans-serif for ALL text
- Style: minimal, elegant — rounded rects (rx=8-14), subtle glows
- Section headers: font-size="10" letter-spacing="2" fill="#A1AABB" uppercase
- Filter IDs MUST be prefixed with `s{N}-`

## COLOR PALETTE (Hefesto Forge)
- fg: #F5F0E8, subtext: #A1AABB, muted: #8B8175
- surface: #12121a, surface2: #1A1A22, line: #3D3529
- orange: #E8850C, green: #4ADE80, yellow: #E8A84C
- red: #C4453A, copper: #C17F59, magenta: #E87040
- cyan: #DEBA87, accent: #E8850C

## SLIDE CONTENT
Title: "{title}"
Concept: {detailed description of what to diagram}

Return ONLY raw SVG markup. Start with <svg, end with </svg>.
```

---

## Commands

```bash
# Serve locally for development
python3 -m http.server 8080     # Then open http://localhost:8080

# Count slides
rg -c '<article class="slide"' index.html

# Verify no remaining <img> tags
rg 'assets/images/' index.html || echo "ALL INLINE"

# Count SVGs
rg -c '<svg' index.html

# Check for bad contrast colors (should return nothing)
rg '#FFE066|#3D3529(?!.*stroke)' index.html || echo "CLEAN"

# Audit all fill colors used
rg -o 'fill="#[0-9a-fA-F]{6}"' index.html | sort | uniq -c | sort -rn
```

---

## Glass Morphism & Depth System

The visual depth comes from layered effects, NOT gradients:

```css
/* Card depth recipe */
.card {
  border: 1px solid var(--line);       /* subtle edge */
  border-radius: 16px;
  background: var(--gray1);            /* solid surface */
  box-shadow:
    0 16px 34px rgba(0, 0, 0, 0.36),  /* drop shadow */
    inset 0 1px 0 rgba(255,255,255,0.03); /* top highlight */
  backdrop-filter: blur(8px);          /* glass effect */
}

/* Ambient glow (background layer) */
.ambient::before {
  border-radius: 999px;
  filter: blur(120px);
  opacity: 0.18;
  background: var(--orange);           /* forge glow */
}
```

**Rule: NO CSS gradients.** Use solid colors + shadows + blur for depth.

---

## Fonts

- **Headings**: `Source Serif 4` (serif, weight 500/700)
- **Body/UI**: `Space Grotesk` (sans-serif, weight 400/500/700)

```html
<link href="https://fonts.googleapis.com/css2?family=Space+Grotesk:wght@400;500;700&family=Source+Serif+4:wght@500;700&display=swap" rel="stylesheet" />
```

---

## Navigation Features

- **Arrow keys** / PageUp/PageDown: navigate slides
- **Space**: next slide
- **Home/End**: first/last slide
- **F key**: toggle focus mode (hides rail sidebar)
- **Touch swipe**: mobile navigation
- **Rail clicks**: jump to module or specific slide dot
- **Animated transitions**: slide-enter/leave with cubic-bezier easing (420ms)
- **localStorage persistence**: saves current slide index, restores on reload

---

## Slide Persistence (localStorage)

The app saves the current slide to `localStorage` on every navigation, and restores it on page load:

```js
// Save on every navigation (inside render function, after updating current)
try { localStorage.setItem("deck-slide", current); } catch (_) {}

// Restore on load (at bottom of app.js)
const saved = Number(localStorage.getItem("deck-slide")) || 0;
render(Math.min(saved, slides.length - 1), { animate: false });
```

This means reloading the browser returns to the same slide. The `try/catch` handles private browsing where localStorage may be unavailable.

---

## Adding Slides to an Existing Deck

When inserting a new slide into an existing deck, you MUST re-index everything downstream:

1. **Insert the `<article class="slide">`** with the correct `data-index`, `data-module`, `data-tone`
2. **Increment `data-index`** on ALL subsequent slides (work highest to lowest to avoid collisions)
3. **Increment kicker numbers** in `<p class="slide-kicker">NN · Module</p>` (highest to lowest)
4. **Update `data-first`** on rail buttons for all modules AFTER the insertion point (+1 each)
5. **Update the counter** in the HTML (`01 / NN`)

**Always work from HIGHEST index down to LOWEST** to avoid double-incrementing.

Use a sub-agent (Task tool) for re-indexing — it's tedious but critical for navigation to work.

---

## SVG Layout Gotchas

### All elements MUST fit inside viewBox (0 0 520 360)

- Max Y for any element: **350** (leave 10px margin)
- Max X for any element: **510** (leave 10px margin)
- **Center diagrams horizontally**: calculate total width of all elements + gaps, then offset = `(520 - totalWidth) / 2`
- Elements that grow in size (like progressive documents) should be **bottom-aligned** — anchor their bases to the same Y line and let them grow upward
- **NEVER** place text labels outside the viewBox — they will be invisible or clipped
- Filter effects (`feGaussianBlur`) add visual bleed — account for ~4px extra around filtered elements
- When using `transform="translate(x, y)"`, all child coordinates are RELATIVE — add translate Y + child Y to get absolute position

### Checklist after editing any SVG:
```
1. Is the highest element > Y=10? (not clipped at top)
2. Is the lowest element < Y=350? (not clipped at bottom)
3. Is the leftmost element > X=10? (not clipped at left)
4. Is the rightmost element < X=510? (not clipped at right)
5. Are elements visually centered in the 520px width?
```

---

## Resources

- **Color source**: `HefestoOpenCode/themes/hefesto.json`
- **Reference implementation**: See [assets/](assets/) for HTML scaffold template
