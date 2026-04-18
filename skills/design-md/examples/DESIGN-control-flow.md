# Design System: Control Flow — if vs switch, for variants, select
**Project ID:** [your-project-id]
**Folder Family:** `control-flow / logic`
**Document Role:** Conceptual explainer with decision guidance

---

## 1. Visual Identity

### Mood Statement
Like standing at a rail junction at dusk — multiple tracks diverging into distance, each lit differently, each leading somewhere distinct. The document doesn't tell you which track to take; it shows you where each one goes. There is no anxiety here, only the clarity of well-drawn options. The visual weight sits on **divergence**, not chaos.

### Visual Metaphor
A **subway junction map** viewed from directly above — clean color-coded lines branching from a central node, each path labeled, no line crossing another without intention. The image implies: *every branch was designed, not accidental.*

### Design DNA
Inherits from the control-flow family: directional energy, geometry as meaning. But this specific document is about **comparison**, not flow — so the geometry shifts from arrows-in-sequence to **parallel lanes** that start at the same point. Siblings like "goroutines" or "channels" get rhythm and repetition; this document gets **divergence and taxonomy**.

---

## 2. Hero Image Brief

**Concept**: A top-down view of three train tracks diverging from a single junction point, each track a different color, extending into mist

**Visual Metaphor**: Rail junction — same origin, different destinations, no hierarchy between paths

**Composition**: Centered origin point with three lines extending asymmetrically — two rightward, one leftward — filling 70% of the frame. Heavy negative space in the top third.

**Color Mood**: Near-black background (#0D0D0D), with three track colors — Electric Amber (#F5A623) for `if`, Cold Cyan (#3DD6F5) for `switch`, Muted Coral (#E8735A) for `select`. Track colors fade toward the edges into near-invisible.

**Texture / Material**: Brushed steel for the tracks. Subtle grain overlay on the entire image (5% opacity). The junction point has a slight glow — not dramatic, just warm.

**Typography in Image**: Document title in uppercase, wide-tracking, Fira Mono — anchored bottom-left. Tag label "CHOOSER" top-right in small caps, same amber as the `if` track.

**What it should feel like**: The moment before choosing — calm, informed, no wrong answer. Like a seasoned engineer pausing at a whiteboard.

**What to avoid**: Flowchart boxes. Pastel backgrounds. Generic "branching tree" clip art. Anything that implies one path is better than another.

---

## 3. Color Palette

- **Midnight Soot** (#0D0D0D) — Primary background. Not pure black — slightly warm to prevent harshness. Sets the high-contrast stage.
- **Electric Amber** (#F5A623) — Primary accent for `if` expressions. The oldest, most instinctive tool; amber implies intuition.
- **Cold Cyan** (#3DD6F5) — Accent for `switch` statements. Structural, systematic, cool. Implies a deliberate architecture.
- **Muted Coral** (#E8735A) — Accent for `select` + concurrency context. Warm but not alarming — signals "this is where things get interesting."
- **Pale Slate** (#C4C9D4) — Body text and supporting annotations. Readable against dark without being clinical white.
- **Dim Rail Gray** (#3A3D45) — Dividers, code block backgrounds, secondary containers. The infrastructure color.

*Why this palette*: Three distinct hues for three control structures — the reader learns to associate color with concept before reading a word. No neutral palette could carry this.

---

## 4. Typography System

**Primary**: `Syne` — geometric, confident, slightly wide. Used for all headings. Its letterforms suggest structure without coldness.
- H1: Syne Bold, 2.5rem, letter-spacing 0.04em, Electric Amber
- H2: Syne Medium, 1.75rem, letter-spacing 0.02em, Pale Slate
- H3: Syne Regular, 1.25rem, Cold Cyan

**Secondary/Prose**: `IBM Plex Sans` — engineered, readable, neutral without being invisible. Body text at 1rem, line-height 1.75, Pale Slate (#C4C9D4).

**Code**: `Fira Mono` — the classic. Code blocks at 0.9rem, syntax-highlighted against Dim Rail Gray backgrounds.

*Why this pairing*: Syne has an architectural feeling — it was designed for posters and navigation. IBM Plex Sans was designed by engineers for engineers. The combination reads: "this was built, not styled."

---

## 5. Component Language

### Structural Elements
- **Section cards**: Sharp corners (0px radius). Thin 1px left border in the accent color of the topic being discussed (Amber for `if`, Cyan for `switch`). Dark background (#181820).
- **Dividers**: 1px horizontal rules in Dim Rail Gray. No decorative elements — the line is the feature.
- **Comparison tables**: 3-column layout. Each column header color-coded to match its control structure. No rounded cells.

### Interactive Elements
- **Buttons/CTAs**: Rectangular, 2px border in accent color, transparent fill. Hover: fill with 15% opacity accent. No border-radius.
- **Tag labels** (like "CHOOSER", "ENTRY"): Small-caps, 0.7rem, border 1px, tight padding, same accent palette.

### Code Blocks
- **Background**: Dim Rail Gray (#3A3D45) with subtle left border in topic's accent color
- **Syntax palette**: Amber for keywords (`if`, `switch`, `for`), Cyan for identifiers, Coral for special cases, Pale Slate for everything else
- **Font**: Fira Mono, 0.875rem, line-height 1.6
- **Padding**: 1.25rem horizontal, 1rem vertical. No border-radius.

### Callout Boxes
- **"Use when"** callout: Left border in Cold Cyan, dark background, IBM Plex Sans Italic, slightly indented
- **"Watch out"** callout: Left border in Muted Coral. Same structure.
- **"Key insight"** callout: Left border in Electric Amber. Bold label, then normal weight body.

---

## 6. Layout Principles

- **Max width**: 820px for content column (reading-optimized, not full-bleed)
- **Grid**: Single column with 48px side padding. No sidebar — this document deserves full focus.
- **Vertical rhythm**: 2rem between components, 4rem between major sections
- **Hero image**: Full-bleed at top, 400px height. Text overlays anchored bottom-left.
- **Comparison sections**: Three equal-width columns when comparing `if / switch / select` side by side. Collapse to accordion on mobile.
- **Code-prose ratio**: 40/60. More prose than code — this is a guide, not a reference sheet.

---

## 7. Stitch Generation Notes

### Atmosphere Phrase
"A high-contrast technical document with a rail-junction aesthetic — three color-coded paths diverging from a shared origin, set against near-black with amber, cyan, and coral accents. Precise, deliberate, no ornamentation."

### Color References
- Primary background: "Midnight Soot (#0D0D0D)"
- `if` accent: "Electric Amber (#F5A623)"
- `switch` accent: "Cold Cyan (#3DD6F5)"
- `select` accent: "Muted Coral (#E8735A)"
- Body text: "Pale Slate (#C4C9D4)"
- Containers: "Dim Rail Gray (#3A3D45)"

### Component Prompt Templates
- "Create a comparison card with sharp corners, 1px left border in Electric Amber, dark Midnight Soot background, and Syne Bold heading"
- "Design a code block with Dim Rail Gray background, Fira Mono font, and keyword highlighting in Electric Amber"
- "Add a 'Use when' callout with Cold Cyan left border, IBM Plex Sans Italic body text, no border-radius"

### What Makes This Document Unique
While siblings in the control-flow family use directional arrows and sequence, this document is fundamentally about **taxonomy and comparison** — three tools, three identities, one decision point. The three-accent color system is unique to this document; no other document in the project should use all three simultaneously.
