# Package Boundary Rules

- Organize by bounded context first, then by layer where that improves ownership clarity.
- Define interfaces near the code that consumes them.
- Keep DTOs and transport schemas out of domain packages.
- Avoid circular dependencies by extracting shared policies into domain services or `pkg/*` primitives only when reuse is real.
- Do not create generic repositories or helpers that erase business intent.
- Prefer explicit package names such as `order`, `inventory`, `promotion`, `postgres`, `redis`, `httpapi`.
