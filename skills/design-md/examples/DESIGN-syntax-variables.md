# Design System: Syntax & Variables — var, :=, const
**Project ID:** [your-project-id]
**Folder Family:** `syntax / variables / basics`
**Document Role:** Entry point — first real encounter with Go's declaration model

---

## 1. Visual Identity

### Mood Statement
A new terminal window, cursor blinking at column 0. Everything is possible, nothing has been typed yet. This is the document where readers make their first real decision in Go — and the design should feel like the first keystroke: deliberate, immediate, slightly exciting. Not intimidating. Not childishly simple. **The aesthetic of a blank slate with intention.**

### Visual Metaphor
A **terminal at the exact moment of first input** — the cursor just appeared, the prompt is fresh, one line of code is being typed but not yet completed. The image suggests: *you are here, and what you write next matters.*

### Design DNA
The syntax family is the foundation — everything else in the project builds from it. This document therefore gets the most **structurally pure** visual identity: clean grid, monospace primary, black-and-white with a single deliberate accent. No complexity, no metaphorical weight. The design says: *start here, understand this, then build.*

Unlike sibling documents in deeper folders (control flow, runtime), this entry point gets **high white space** and **low visual density** — the reader shouldn't feel overwhelmed before they've written their first `:=`.

---

## 2. Hero Image Brief

**Concept**: A terminal window photographed slightly from the side — not a screenshot, but the actual physical screen — showing three lines of Go code, one variable declaration per line, cursor blinking after the third

**Visual Metaphor**: The blank terminal as creative tool — not a console of errors, but the starting point of something

**Composition**: Terminal window occupies left 65% of frame. Right 35%: dark gray, out of focus. The three lines of code are center-vertically aligned. Cursor glow provides the only warm light.

**Color Mood**: Matte black (#111111) for the terminal body. Green-white phosphor text (#E8F4E8) for the code. Single cursor glow in Warm Green (#4ADE80). The background behind the terminal: a deep, desaturated gray (#1A1A1A).

**Texture / Material**: Slight screen reflection in the terminal glass — not clean, slightly used. Matte bezel. No desk surface visible.

**Typography in Image**: The three code lines in Fira Code with ligatures enabled. Document title "SYNTAX & VARIABLES" in small-caps, wide tracking, anchored bottom-right — Muted Slate. Tag "ENTRY" top-right in Warm Green.

**What it should feel like**: That specific anticipation of opening a terminal to start something new. Focused. Slightly quiet. Ready.

**What to avoid**: Colorful, toy-like illustration. White backgrounds with pastel blobs. Any image that looks "beginner" in a condescending way. Also avoid terminal-as-error aesthetics (red text, crash output).

---

## 3. Color Palette

- **Terminal Black** (#111111) — Primary background. Warmer than pure black — the color of a high-quality monitor at minimum brightness.
- **Phosphor White** (#E8F4E8) — Primary text. Slightly green-tinted white, referencing CRT terminal phosphor. Not clinical — slightly alive.
- **Warm Green** (#4ADE80) — Single accent. Used for `:=` (the short declaration), the cursor, and key highlights. Implies "this works, this is idiomatic."
- **Slate Blue** (#6B7FA8) — Secondary accent. Used for `var` and `const` — the more formal, deliberate declarations. Implies structure and permanence.
- **Dim Ember** (#3A3A3A) — Code block backgrounds and card surfaces. Barely distinguishable from Terminal Black at a glance — the difference is felt, not seen.
- **Fog Gray** (#9CA3AF) — Body text for prose sections. Readable against Terminal Black without competing with Phosphor White headings.

*Why this palette*: Green-and-black is the original coding aesthetic — this document earns it by being the entry point to Go syntax. The green/blue split between `:=` and `var`/`const` encodes the semantic difference visually before the prose explains it.

---

## 4. Typography System

**Primary**: `Fira Code` — but used for headings, not just code. Its ligature support (→, :=, !=) means structural characters render as meaningful glyphs in headers. H1 at 2.25rem uses the `:=` ligature as a visual motif.
- H1: Fira Code Regular, 2.25rem, Phosphor White, `:=` motif in Warm Green
- H2: Fira Code Regular, 1.5rem, Phosphor White
- H3: Fira Code Light, 1.125rem, Fog Gray

**Secondary/Prose**: `Nunito` — friendly, slightly rounded, optimized for readability. Not childish — its roundness here means *approachable*, not *simple*. Body at 1rem, line-height 1.8, Fog Gray.

**Code**: `Fira Code` with ligatures, 0.9rem. In code blocks: `:=` in Warm Green, `var`/`const` in Slate Blue, identifiers in Phosphor White, everything else in Fog Gray.

*Why this pairing*: Fira Code for both headings and code creates visual unity — the whole document feels like it was built inside the terminal. Nunito's warmth provides breathing room in prose without breaking the terminal aesthetic.

---

## 5. Component Language

### Structural Elements
- **Section cards**: 6px radius (slightly rounded — approachable). Background: Dim Ember (#3A3A3A). No border — the background contrast is enough.
- **Comparison table** (var vs := vs const): Dark header row in Terminal Black with Phosphor White text. Each column subtly color-coded: `var` header in Slate Blue tint, `:=` in Warm Green tint, `const` in Fog Gray tint.
- **Dividers**: Single 1px in #2A2A2A — barely visible, just enough to segment.

### Interactive Elements
- **Links**: Warm Green, no underline at rest. Underline on hover. Simple.
- **"Try it" prompts**: Small, rounded pill buttons. Background: Warm Green at 15% opacity. Border: 1px Warm Green. Text: Warm Green. Hover: 30% opacity fill.

### Code Blocks
- **Background**: Dim Ember (#3A3A3A)
- **Border**: None — the background contrast handles separation
- **Syntax**: Warm Green for `:=`, Slate Blue for `var`/`const` keywords, Fog Gray for types, Phosphor White for identifiers
- **Padding**: 1.25rem. The code needs room, not luxury.
- **Line numbers**: Fog Gray, 0.75rem, right-aligned in 2rem column. Subtle, functional.

### Callout Boxes
- **"Zero-value" callout**: Slate Blue left border. Body in Nunito, explains Go's zero-initialization. No icon.
- **"Scope"** callout: Warm Green left border. The most important concept — green signals "pay attention."
- **"Common mistake"** callout: No border. Italic Nunito. Slightly lower opacity (90%). Mistakes are acknowledged quietly, not dramatized.

---

## 6. Layout Principles

- **Max width**: 900px — wider than the panic/defer document. Entry content benefits from side-by-side examples.
- **Two-column zones**: When comparing `var` vs `:=` side by side, a 50/50 split with Dim Ember cards. Each card shows one declaration style.
- **Vertical rhythm**: 1.5rem between components, 3.5rem between major sections. Tighter than deeper documents — this is a brisk, focused read.
- **Hero image**: Full-bleed, 380px. Slightly shorter than runtime docs — entry content shouldn't feel heavy.
- **Code/prose ratio**: 35/65. More prose — the reader needs context, not just patterns.
- **First 3 scrolls rule**: The most important concept (`:=` is idiomatic, `var` is for explicit typing, `const` is immutable) should be fully visible without scrolling.

---

## 7. Stitch Generation Notes

### Atmosphere Phrase
"A terminal-aesthetic entry document — Terminal Black background, phosphor-white and warm-green typography, Fira Code used for both headings and code, with a friendly Nunito prose voice. Approachable without being naive."

### Color References
- Background: "Terminal Black (#111111)"
- Primary text: "Phosphor White (#E8F4E8)"
- `:=` accent: "Warm Green (#4ADE80)"
- `var`/`const` accent: "Slate Blue (#6B7FA8)"
- Code backgrounds: "Dim Ember (#3A3A3A)"
- Body text: "Fog Gray (#9CA3AF)"

### Component Prompt Templates
- "Create a comparison card in Dim Ember background, 6px radius, Fira Code heading, showing a single Go declaration type with syntax highlighted in Warm Green"
- "Design a two-column layout with Terminal Black background, each column an Dim Ember card — left column shows `var`, right shows `:=`"
- "Add a scope callout with 2px left border in Warm Green, Nunito body text, no icon, no border-radius"

### What Makes This Document Unique
This is the only document in the project that uses Fira Code for *both* headings and code — creating a fully terminal-native visual identity. The Warm Green / Slate Blue semantic split (idiomatic vs formal) is unique to this document's teaching goal. No other document encodes syntactic meaning directly in accent color choice.
