# Visual Kinds

Use `visual_kind` to choose the layout family, but choose from the **teaching job**, not from convenience.

## `state_trace`

Use when the reader must see state evolve over time.

Best for:

- sorting passes
- pointer movement
- BFS/DFS frontier changes
- same input, different state -> different result

This is often the best default for DSA visuals when a card wall would feel flat.

## `boundary_map`

Use when the reader must see control, ownership, capability, or dependency direction across a boundary.

Best for:

- ports and adapters
- driver vs driven
- team boundary / system boundary
- request ownership vs infrastructure responsibility

## `workflow_map`

Use for sequence, lifecycle, escalation flow, or handoff.

Best for:

- migration rollout flow
- incident escalation path
- protocol or request lifecycle

## `compare_card`

Use for side-by-side contrast when the core question is truly "A vs B."

Best for:

- REST vs RPC
- mutex vs channel
- before vs after

If the compare needs state change or a richer narrative arc, prefer `state_trace` or a purpose-built visual instead.

## `decision_map`

Use when the reader is trying to choose between options.

Best for:

- `var` vs `:=`
- interface vs concrete
- binary-search variant chooser

## `router_map`

Use for hubs and route-selection docs.

Best for:

- learning path map
- symptom -> module router
- subtree overview

## `taxonomy_card`

Use for grouped families or classification.

Best for:

- diagram families
- DSA pattern families
- glossary category overviews

## `mental_model_card`

Use when the main job is correcting the reader's intuition.

Best for:

- stack vs heap
- slice semantics
- nil interface trap

## `api_family_map`

Use for function families, helper clusters, or standard-library choice maps.

Best for:

- `strings`
- `strconv`
- `fmt`
- collection helper families
