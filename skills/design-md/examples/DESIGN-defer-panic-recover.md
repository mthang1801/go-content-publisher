# Design System: Defer / Panic / Recover
**Project ID:** [your-project-id]
**Folder Family:** `runtime / crash-boundary`
**Document Role:** Conceptual explainer — when the happy path diverges

---

## 1. Visual Identity

### Mood Statement
The moment before a system catches itself — not catastrophe, but the sharp intake of breath before recovery begins. This document lives in the space between error and handling, and its visual language should carry that tension without resolving it prematurely. The aesthetic is **clinical urgency**: a hospital corridor at 3am, fluorescent and purposeful, not panicked but not relaxed.

### Visual Metaphor
A **circuit breaker mid-trip** — photographed at the exact instant the breaker arm is between closed and open. The mechanism is exposed, the physics visible, the outcome not yet determined. Defer is the spring; Panic is the fault current; Recover is the breaker catching.

### Design DNA
Runtime family documents share high-contrast aesthetics and a sense of consequence. But while "goroutines" (a sibling) gets energy and parallelism, this document gets **threshold and recovery**. The palette cools toward resolution — tension enters red, exits in measured slate.

---

## 2. Hero Image Brief

**Concept**: A macro photograph of an electrical circuit breaker, frozen mid-trip — the arm visible, the arc of the trip mechanism caught at its apex

**Visual Metaphor**: The circuit breaker as the body's nervous system under stress — not failing, but *protecting*

**Composition**: Extreme close-up, slightly off-center left. The breaker mechanism fills 60% of frame. Right side: deep shadow fading to black. A hairline of red light traces the arc of the trip mechanism.

**Color Mood**: Dominant near-black (#0A0A12), interrupted by a single line of Fault Red (#E63946). Recovery Blue-Gray (#4A7FA5) visible in the distant background — the color of resolution, not yet reached.

**Texture / Material**: Metal and plastic — actual circuit breaker materials. Subtle lens blur on foreground edges. No digital illustration.

**Typography in Image**: "DEFER / PANIC / RECOVER" in tight monospace, small, anchored top-right. A small tag "RUNTIME" in Fault Red, 0.65rem.

**What it should feel like**: Like reading the label on a fire extinguisher — calm, technical, designed for the worst moment. You trust it.

**What to avoid**: Red warnings or skulls (too alarmist). Explosion imagery. Any visual that implies the system *fails* rather than *recovers*.

---

## 3. Color Palette

- **Deep Space Black** (#0A0A12) — Background. Slightly blue-black, not pure black. The color of a server room at night.
- **Fault Red** (#E63946) — Used *only* for `panic` contexts. High-contrast accent. Appears sparingly — its rarity is its power.
- **Recovery Blue-Gray** (#4A7FA5) — Used for `recover` and `defer` contexts. The color that follows red. Calm, procedural, trustworthy.
- **Dim Console** (#1E1E2E) — Code block backgrounds, secondary cards. The terminal color.
- **Muted Phosphor** (#A8B2C0) — Body text. Slightly blue-tinted gray — evokes old CRT monitors. Readable, slightly eerie.
- **Hairline Gray** (#2D2D3D) — Borders, dividers. Visible only in context.

*Why this palette*: The color story IS the document's story — red enters (panic), blue resolves it (recover), black contains both. The reader experiences the arc of the mechanism through color alone.

---

## 4. Typography System

**Primary**: `Space Grotesk` — wait, this is on the avoid list. Use `Barlow Condensed` instead — tall, narrow, slightly industrial. Perfect for headers in a document about system internals.
- H1: Barlow Condensed SemiBold, 3rem, letter-spacing 0.06em, Muted Phosphor
- H2: Barlow Condensed Medium, 2rem, Recovery Blue-Gray
- H3: Barlow Condensed Regular, 1.4rem, Muted Phosphor with Fault Red left border

**Secondary/Prose**: `Source Serif 4` — a serif in a technical document creates unexpected gravitas. Body text at 1.05rem, line-height 1.8, Muted Phosphor. The slight oldness of a serif fits a document about error recovery — it implies *experience with failure*.

**Code**: `Roboto Mono` — Not Inter (body), but Mono is fine here. 0.875rem. Keywords (`panic`, `recover`, `defer`) highlighted in Fault Red.

*Why this pairing*: Condensed sans for structure + old-school serif for prose = a document that feels both engineered and considered. Like a safety manual written by someone who has seen the worst.

---

## 5. Component Language

### Structural Elements
- **Section cards**: Very slight radius (4px — barely perceptible). Background Dim Console (#1E1E2E). Left border 2px in Recovery Blue-Gray. When the section discusses `panic`, the left border switches to Fault Red.
- **Dividers**: Hairline Gray (#2D2D3D), 1px. Below major sections only.
- **Execution order diagrams**: Numbered list with large counters (2rem, Barlow Condensed) in Recovery Blue-Gray. The counter IS the design — no extra decoration.

### Interactive Elements
- **Links**: Recovery Blue-Gray, underlined only on hover. No decoration at rest.
- **Key term highlights**: Inline code in Fault Red when `panic`, Recovery Blue-Gray when `defer`/`recover`. Color is semantic, not decorative.

### Code Blocks
- **Background**: Dim Console (#1E1E2E)
- **Border**: 1px left, color reflects the mechanism shown (`panic` → Fault Red, `defer`/`recover` → Recovery Blue-Gray)
- **Syntax**: `panic` and `recover` keywords in Fault Red. Deferred functions in Recovery Blue-Gray. Everything else in Muted Phosphor.
- **Padding**: 1.5rem. Generous — this code needs space to breathe.

### Callout Boxes
- **"Cleanup order"** callout: Numbered, Recovery Blue-Gray. Tight grid layout — order matters.
- **"Danger zone"** callout: Fault Red left border, 2px. Body in Source Serif 4 Italic. Used only when misuse causes silent failures.
- **"Key insight"** callout: No border. Just increased left padding (2rem) and slightly larger font. The restraint is the signal.

---

## 6. Layout Principles

- **Max width**: 760px. Narrower than siblings — this is a focused, deep document, not a reference sweep.
- **Vertical rhythm**: 2.5rem between components, 5rem between phases (Setup → Panic → Recovery)
- **Hero image**: Full-bleed, 480px. Dark enough that text overlays read clearly at 100% opacity.
- **Execution flow diagrams**: Horizontal, numbered, with connecting lines in Hairline Gray. Flow reads left-to-right (setup → defer stack → panic → recover → continue).
- **Code/prose ratio**: 50/50. Both matter equally here — the code shows the pattern, the prose explains the mechanism.

---

## 7. Stitch Generation Notes

### Atmosphere Phrase
"A clinical, high-contrast technical document with a circuit-breaker aesthetic — deep space black with fault-red and recovery-blue-gray accents, Barlow Condensed headers, and Source Serif body text. Tension enters red, resolves in blue."

### Color References
- Background: "Deep Space Black (#0A0A12)"
- Panic accent: "Fault Red (#E63946)"
- Defer/recover accent: "Recovery Blue-Gray (#4A7FA5)"
- Code backgrounds: "Dim Console (#1E1E2E)"
- Body text: "Muted Phosphor (#A8B2C0)"

### Component Prompt Templates
- "Create a code block with Dim Console background, 2px left border in Fault Red, Roboto Mono font with panic keyword highlighted in Fault Red"
- "Design a section card with barely-perceptible 4px radius, Dim Console background, Recovery Blue-Gray left border, and Barlow Condensed SemiBold header"
- "Add a cleanup-order callout with numbered items, Recovery Blue-Gray numbering, tight grid layout — no decorative elements"

### What Makes This Document Unique
The color system is **mechanistic, not stylistic** — Fault Red and Recovery Blue-Gray map directly to `panic` and `recover`. No other document in the project uses this semantic color-to-concept mapping. The use of Source Serif 4 for body text (a serif in a technical context) creates deliberate gravitas unique to this document's stakes.
