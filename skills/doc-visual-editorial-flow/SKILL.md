---
name: doc-visual-editorial-flow
description: >
  Use when a documentation visual should feel like an editorial explainer rather
  than a card grid: architecture boundaries, protocol handoffs, system flows,
  central-metaphor diagrams, or when the user references sketch-like Facebook /
  3DongCode style visuals with organic connectors, sticky-note nodes, and a
  strong visual story.
---

# doc-visual-editorial-flow

> **Vị trí trong workflow**: Đây là **Ưu tiên 3** trong visual strategy (sau `nano-banana-2` / `nano-banana` và Research-Adapt). Chỉ dùng skill này khi Ưu tiên 1 (`nano-banana-2` rồi `nano-banana`) và Ưu tiên 2 (research-adapt từ web) không phù hợp cho teaching job.

Use this skill when the current visual grammar is too rigid.

Do **not** use this skill as the default path for every "beautiful visual" request.
If the target quality is closer to the image-native, highly polished family already seen in `assets/dsa/images`, prefer `nano-banana-2` or `nano-banana` first, or research-adapt from web sources. Use this skill only when a scripted, deterministic editorial diagram is still the right medium and nano-banana cannot produce the needed level of structural control.

This skill is for visuals that should feel:

- editorial, not dashboard-like
- natural to the topic, not template-driven
- centered around one strong metaphor
- easy to scan because the eye is guided by composition, not by a wall of cards

This is the right fit for:

- Ports & Adapters / hexagonal architecture
- request or protocol flow
- multi-actor handoff
- system boundary / ownership split
- product or architecture diagrams where a central concept should anchor the page

Do not use this skill for:

- plain taxonomies or router maps where an image-first explainer from `nano-banana-2` or a simple diagram fallback already teaches the point clearly
- simple compare cards
- screenshot-style visuals
- photoreal or illustration-heavy hero art
- image-native explainers whose quality bar depends on rich illustration craft more than deterministic layout control

## What to read

Read only what you need:

- `references/quality-bar.md` when the visual is already directionally right but still feels rough
- `references/icon-registry.ts` when the diagram needs recognizable actor, platform, infra, or architecture notation
- `design-md` when the folder still lacks a strong visual brief

## Core Style Cues

The target visual language is an **editorial sketch explainer**:

- one strong center of gravity
- satellites around it, each with a clear role
- connectors that read like movement or causality
- small labels / sticky-note nodes instead of repeated large cards
- generous whitespace
- a light legend when the reader needs quick decoding
- a hand-drawn or lightly sketched feeling without becoming sloppy

Good outputs should remind the reader of a thoughtful teaching board, not a slide deck full of UI cards.

The quality bar is higher than "the idea is correct."
This skill is only done when the visual is also:

- sharp in typography
- calm in spacing
- legible at a glance
- specific in iconography
- free of accidental overlap or muddy focus

## Workflow

### 1. Start from the teaching job

Write down:

- what exact question the visual should answer
- what the central concept is
- what belongs at the edge
- what the reader should notice first, second, third

If there is no central teaching question, do not draw yet.

### 2. Inspect repo-native DNA first

Before looking outside:

- inspect the strongest local references in the repo
- preserve any visual DNA that already works, especially from `assets/dsa/images`
- only expand outward if the existing local grammar is insufficient

Never replace a strong repo-native style with something flatter or more generic.

### 3. Research references on the web

Research `2-5` references that match the same teaching job.

When studying each reference, extract:

- center vs edge composition
- reading path
- connector style
- label density
- use of whitespace
- whether the visual uses a legend, grouping, or color coding

Do **not** copy:

- exact layout
- exact text
- shape placement one-to-one
- distinctive palette or decoration in a way that makes the source obvious

The point is to capture **idea-space**, not clone expression.

### 4. Build a visual brief

Write a compact internal brief with:

- teaching question
- metaphor
- center anchor
- edge actors / systems
- connection semantics
- palette direction
- notation style
- legend need or no-legend decision

If the brief still sounds like "some cards with bullets," stop and strengthen it.

Also state:

- what should feel sharp vs loose
- where icon recognition matters
- where text must be minimized
- what should still be understandable when the image is viewed small

### 5. Choose implementation path

Preferred order:

1. If the visual is really an image-native explainer, start with `nano-banana-2` or `nano-banana` using repo-native references plus web references.
2. If a repo-native renderer can express the composition without flattening it, use it.
3. If not, create a bespoke SVG/PNG asset instead of forcing the idea into the wrong template.
4. If needed, extend tooling after the visual language is already clear.

Important:

- do not force editorial-flow visuals into a card wall
- do not keep using a generic renderer once it starts erasing the metaphor

### 6. Draw with editorial composition rules

Use these rules while composing:

- the central concept should be visually dominant
- edge actors should feel peripheral but connected
- connectors should show direction and responsibility
- callouts should be short and positioned where the eye naturally pauses
- labels should decode, not duplicate prose
- keep empty space on purpose; crowding kills the editorial feel

Typography rules:

- title hierarchy must be obvious even if the body text is blurred
- do not let decorative blobs or shapes compete with the title
- never allow annotation text to collide with arrows, nodes, or other labels
- prefer fewer words with stronger placement over full-sentence boxes everywhere

Icon rules:

- start from `references/icon-registry.ts` so icon choices stay consistent across the repo
- when a real-world tool or platform matters, use a recognizable icon or redraw-inspired glyph
- inspect official or stable public references first, then redraw or adapt for the repo visual
- treat emoji in the registry as semantic shorthand, not as a hard requirement to paste emoji into the final PNG
- prefer bespoke redrawn glyphs for publication-quality visuals, using the registry mainly to normalize meaning and coverage
- icons must help recognition fast; they should not feel like clip-art pasted on top
- if the icon adds noise, remove it

### 7. Redraw from scratch

The final asset must be a new work.

Allowed:

- same idea family
- same teaching job
- similar metaphor class

Not allowed:

- recolor-and-ship
- trace-over with minimal changes
- swap text into an obvious copied composition

### 8. Verify

Before finalizing, check:

- does the eye know where to start?
- is there a visible center of gravity?
- does the diagram feel natural to the topic?
- if all text is blurred, does the composition still teach roughly the right thing?
- does it still feel like this repo, not a pasted foreign style?
- are the fonts, weights, and line breaks sharp rather than merely acceptable?
- do icons increase recognition without making the diagram noisier?
- is any text colliding, crowding, or visually "buzzing" against connectors or shapes?
- does the image survive a 50% zoom-out without turning into mush?

### 9. Refine before shipping

Do not stop at the first render that is "basically right."

Run at least one refinement pass focused on craft:

- remove overlaps
- rebalance whitespace
- tighten wording
- improve typography hierarchy
- replace generic symbols with stronger, topic-true iconography where appropriate
- simplify any node that still reads like a mini-card instead of an editorial label

If the image still feels rough, do another pass.
This skill assumes iteration, not one-shot output.

## Relationship to Other Skills

- Use `design-md` if the folder needs a stronger visual direction first.
- Use `nano-banana-2` / `nano-banana` as the **first choice** for high-quality image-native explainers.
- Use **Research-Adapt** (search web → adapt via nano-banana-2) as the **second choice** when good visuals already exist online.
- Use this skill as the **third choice** when the user explicitly wants a richer editorial flow or when the visual needs structural control that nano-banana cannot provide.
- `doc-visual-png` is **deprecated** — do not use it for new visual assets.

## Quick Heuristic

Use this skill when the right visual answer sounds like:

- "put the idea in the middle, then show who talks to it"
- "show the request crossing the boundary"
- "make it read like a teaching poster"
- "this should feel organic, not system-generated"
