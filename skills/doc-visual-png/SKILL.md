---
name: doc-visual-png
description: >
  Deprecated skill kept only for legacy reference. Do not use for new doc
  visuals; prefer `nano-banana-2`, then `nano-banana`, then
  `ai-image-generation`, with flow/diagram fallback when image-first is not the
  right fit.
---

# doc-visual-png

> ⛔ **DEPRECATED — KHÔNG SỬ DỤNG CHO VISUAL ASSET MỚI.**
> Skill này và `render_visual.py` **KHÔNG còn được sử dụng** để tạo visual asset cho tài liệu kỹ thuật. Output dạng card-show / card-grid đã bị **loại bỏ hoàn toàn** khỏi workflow `create-doc.md` và R6.
>
> **Thay thế bằng:**
> 1. `nano-banana-2` / `nano-banana` (ưu tiên cao nhất) — image-native generation chất lượng cao
> 2. Research-Adapt — tìm hình ảnh từ web, adapt qua `nano-banana-2`; nếu cần model/capability khác thì mở sang `ai-image-generation`
> 3. `doc-visual-editorial-flow` — khi cần editorial sketch
>
> **Mọi asset hiện tại ở dạng card-show phải được thay thế.** Xem R6 và `create-doc.md` Bước 3 để biết thứ tự ưu tiên mới.

---

*Nội dung bên dưới giữ lại cho mục đích tham khảo cấu trúc spec/JSON cũ. KHÔNG dùng để tạo asset mới.*

Use this skill to create deterministic `.png` teaching visuals locally, but do not reduce the job to "render some cards."

This skill is **not** the default choice for every high-craft documentation image.
If the target quality is closer to the repo's stronger image-native family such as `assets/dsa/images`, or the article needs a richer and more natural visual metaphor than a scripted renderer can comfortably provide, prefer `nano-banana-2` or `nano-banana` first.

The default target is a **visual-explainer**:

- easy to enter for non-expert readers
- has a clear reading path
- shows state, boundary, motion, or comparison in a way prose alone would flatten
- uses spatial hierarchy so the eye knows where to start and what matters most

If the repo already has a strong local visual language for the same family, preserve that first.
`assets/dsa/images` is a good example of repo-native quality to learn from, not flatten.

Do not use this skill for:

- Mermaid practice blocks
- screenshots of real UI
- photorealistic art
- illustration-first hero art with no teaching job
- editorial sketch-flow posters with a strong center anchor and organic connector language
- image-native explainers where illustration craft matters more than deterministic diagram structure

For those, use a more appropriate skill or workflow.

## What to read

Read only what you need:

- `references/presets.md` when choosing visual tone and palette family
- `references/visual-kinds.md` when choosing the layout family

If the doc or folder needs a stronger visual identity first, use `design-md` before this skill.
If the visual clearly wants a poster-like editorial explainer, use `doc-visual-editorial-flow` instead.

## Workflow

### 1. Confirm the image belongs

Use this skill when the reader would benefit from a visual artifact such as:

- a state trace or pass-by-pass explainer
- a boundary / ownership / contract map
- a flow explainer with handoff or escalation
- a before-after or same-input-different-result comparison
- a router / taxonomy / chooser-like decision visual
- a mental-model correction visual

Do not ask whether the section is literally named `VISUAL`.
Ask whether a PNG would make the section easier to understand, remember, or trust.

### 2. Research first, then redraw

Before choosing layout, research the visual idea-space.

Default approach:

- inspect the best repo-native references first, especially in the same family
- find `2-5` references on the web that match the same teaching job
- study metaphor, pacing, hierarchy, grouping, and how the eye moves
- extract only the *idea-space*, not the expression
- then redraw a brand-new asset for this repo

Hard rules:

- do not overwrite a strong existing repo-native style with something more generic
- do not copy reference text
- do not mirror the layout too literally
- do not just recolor someone else's composition
- do not skip topic understanding because the reference "already looks right"

The output must be **a new visual based on the idea**, not a cloned reference.

### 3. Choose the natural visual metaphor

Do not start from "which template do I have?"
Start from "what should the reader *feel happening* in this topic?"

Examples:

- sorting / algorithm state -> pass-by-pass, progression, storyboard, state trace
- architecture boundary -> lane map, switchboard, control-vs-capability split
- process / incident -> direction, handoff, escalation, checkpoint
- compare -> before-after or same-input-different-state
- routing / taxonomy -> hierarchy or routing logic can be right, but avoid banned card-wall/card-grid patterns

If the same layout would look equally plausible for five unrelated topics, it is probably too generic.

### 4. Choose `content_type`

Pick one:

- `tech`
- `glossary`
- `product`
- `game`
- `career`

This controls the preset palette and typographic tone.

### 5. Choose `visual_kind`

Pick the layout family that matches the teaching job:

- `state_trace`
- `boundary_map`
- `workflow_map`
- `compare_card`
- `decision_map`
- `router_map`
- `taxonomy_card`
- `mental_model_card`
- `api_family_map`

If unsure:

- state evolution, passes, progression -> `state_trace`
- boundary, driver-vs-driven, ownership split -> `boundary_map`
- lifecycle, escalation, sequence of steps -> `workflow_map`
- X vs Y, before vs after -> `compare_card`
- chooser logic -> `decision_map`
- hub or route-selection -> `router_map`
- grouped families -> `taxonomy_card`
- intuition correction -> `mental_model_card`
- helper/library surface -> `api_family_map`

`compare_card`, `decision_map`, `router_map`, and `taxonomy_card` are useful, but they are **not** the default answer for everything.

### 6. Write the brief before the spec

Before JSON, write a compact internal brief:

- teaching question: what exact question should the image answer?
- reading path: where should the eye start, move, and end?
- metaphor: what natural visual language fits the topic?
- state / boundary / comparison: what must visibly change?
- hierarchy: what is primary, secondary, tertiary?

If the brief sounds like "a few cards with bullets," the brief is probably too weak.

### 7. Write the spec

Keep copy short enough to scan inside an image, but let the structure carry the teaching job.

Rules:

- one visual should teach one main insight
- use fewer panels if that gives stronger hierarchy
- prefer stateful labels, transitions, and focused callouts over more bullets
- cards are allowed, but not required
- if the visual needs richer pacing, use `state_lines`, `core`, `side`, and lane labels instead of forcing everything into one flat panel list

Example `state_trace`:

```json
{
  "content_type": "tech",
  "visual_kind": "state_trace",
  "title": "Bubble Sort",
  "subtitle": "Watch local swaps turn into a sorted suffix, one pass at a time.",
  "eyebrow": "Sorting / State Trace",
  "panels": [
    {
      "heading": "Pass 1 starts with local disorder",
      "tag": "Setup",
      "body": "The largest value is still free to drift right.",
      "state_lines": [
        "start: [5, 3, 8, 1, 2]",
        "5>3 swap -> [3, 5, 8, 1, 2]",
        "8>1 swap -> [3, 5, 1, 8, 2]",
        "8>2 swap -> [3, 5, 1, 2, 8]"
      ],
      "bullets": [
        "Only adjacent inversions are repaired",
        "But the pass still locks one global fact"
      ]
    },
    {
      "heading": "Payoff",
      "tag": "Invariant",
      "body": "After one full pass, the largest element of the unsorted region is fixed at the end.",
      "state_lines": [
        "suffix locked: [..., 8]"
      ]
    }
  ],
  "output_path": "assets/dsa/images/example-state-trace.png"
}
```

Example `boundary_map`:

```json
{
  "content_type": "tech",
  "visual_kind": "boundary_map",
  "title": "Ports & Adapters",
  "subtitle": "The core owns decisions. Technology stays at the boundary.",
  "eyebrow": "Architecture / Boundary Map",
  "left_label": "Driver side",
  "right_label": "Driven side",
  "core": {
    "heading": "Application Core",
    "tag": "CORE",
    "body": "Use cases and policy live here.",
    "bullets": [
      "Inbound port receives intent",
      "Outbound port requests capability",
      "Business rules stay framework-agnostic"
    ]
  },
  "panels": [
    {
      "side": "left",
      "heading": "REST Controller",
      "tag": "Driver",
      "bullets": ["Translate HTTP to a use-case command"]
    },
    {
      "side": "left",
      "heading": "CLI Command",
      "tag": "Driver",
      "bullets": ["Starts the same use case from batch entry"]
    },
    {
      "side": "right",
      "heading": "Repository Adapter",
      "tag": "Driven",
      "bullets": ["Implements persistence port"]
    },
    {
      "side": "right",
      "heading": "Payment Adapter",
      "tag": "Driven",
      "bullets": ["Implements payment capability outside the core"]
    }
  ],
  "output_path": "assets/system-design/images/example-boundary-map.png"
}
```

Use `palette_overrides` when the workflow or `design-md` brief has already locked a document-specific Visual DNA. Do not fall back to a generic preset color story when the article needs a distinct visual identity.

### 8. Render the PNG

Run:

```bash
python3 skills/doc-visual-png/scripts/render_visual.py --spec /path/to/spec.json
```

The script writes the final `.png` to `output_path` and creates parent directories if needed.

### 9. Embed it in the doc

Use the normal markdown image pattern:

```md
![Bubble Sort — pass-by-pass explainer](../images/bubble-sort-passes.png)
```

The image should lead the explanation.
The prose should sharpen the takeaway, not repeat every label.

### 10. Verify

Minimum verification:

- the script exits successfully
- the `.png` file exists
- the relative image path in the doc resolves
- the image still matches the section's teaching question after the doc edit
- the image has a readable path for the eye, not just a wall of equally weighted cards
- if you researched web references, the final asset is clearly a redraw, not a near-copy

## Notes for doc workflow

- Prefer `research -> brief -> spec -> render`, not `spec first, metaphor later`.
- Keep Mermaid or code blocks when readers need executable syntax or practice.
- Use this skill for the visual-explainer layer, not the executable layer.
- In glossary docs, `COMPARE` is often the right place for the image.
- In hub `README.md`, `router_map` is useful, but not every hub should look like a chooser.
