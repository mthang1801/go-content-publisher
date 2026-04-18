# Design-MD Skill — Context-Aware Visual Direction System

## Install

```bash
npx skills add google-labs-code/stitch-skills --skill design-md --global
```

## What Changed (vs. original)

The original skill produced a consistent DESIGN.md template across all documents. This refactored version treats every document as a **distinct visual artifact** — where the folder structure, topic family, and content role directly shape the visual direction and PNG brief.

**Key additions**:
- **Phase 0: Read the Room** — infer Design DNA from folder/topic context before touching tokens
- **Folder Family → Visual DNA table** — predefined but extensible conceptual directions per topic category
- **Hero Image Brief** — a required creative brief per document specifying composition, metaphor, color mood, and what to avoid
- **Anti-repetition rules** — explicit guidance to vary palette, font, metaphor across unrelated documents
- **Quality bar checklist** — self-check before finalizing any internal visual brief

## Skill Structure

```
design-md/
├── SKILL.md                          — Core instructions & workflow
├── examples/
│   ├── DESIGN-syntax-variables.md    — Internal brief example: terminal aesthetic, Fira Code headings
│   ├── DESIGN-control-flow.md        — Internal brief example: 3-accent rail-junction metaphor
│   └── DESIGN-defer-panic-recover.md — Internal brief example: circuit-breaker, Fault Red/Recovery Blue
└── README.md                         — This file
```

## Design Philosophy

Every document tells a different story. The image should feel designed *only* for that document — not a reskin of a template.

This skill now works as an **internal briefing step**:
- read the content
- synthesize visual direction
- hand off a PNG / hero image brief
- apply the brief directly in doc refactor and asset generation

It does **not** imply creating a committed `DESIGN.md` inside `assets/` unless the user explicitly asks for a persisted design document.

- Folder families define **Design DNA** (inherited conceptual direction)
- Subfolders **inherit and diverge** from their parent
- Sibling documents share a palette family but differ in accent and composition
- Cross-family documents are visually distinct — different metaphor, palette, voice
- The folder hierarchy is a design hierarchy

## Example Output Comparison

| Document | Metaphor | Primary Font | Key Accent |
|---|---|---|---|
| Syntax & Variables | Terminal cursor | Fira Code (headings + code) | Warm Green #4ADE80 |
| Control Flow | Rail junction | Syne + IBM Plex Sans | Electric Amber #F5A623 |
| Defer/Panic/Recover | Circuit breaker | Barlow Condensed + Source Serif 4 | Fault Red #E63946 |
