# Hooks

These hooks are optional local automation helpers for agents and developers.

## Current Hook Set
- `pre-context-loader.sh`: validates that core project state files exist before deeper work starts.
- `pre-layer-boundary-check.sh`: lightweight content scan for obvious forbidden framework leaks.
- `post-gofmt-check.sh`: warns when Go files are not formatted.
- `post-go-test-check.sh`: runs a targeted or broad test command after implementation work.

## Notes
- Keep hooks fast and safe.
- Hooks should fail loudly on missing prerequisites, but avoid destructive behavior.
- As the sample codebase grows, replace content scans with stronger package/import checks.
