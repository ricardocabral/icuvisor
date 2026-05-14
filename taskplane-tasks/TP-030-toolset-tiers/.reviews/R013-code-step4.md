# R013 code review — Step 4: `icuvisor_list_advanced_capabilities`

Verdict: **APPROVE**

## Findings

No blocking findings.

The R012 catalog-drift issue is fixed: `icuvisor_list_advanced_capabilities` is now present in the adversarial static catalog as a read tool, and the full repository test suite passes. The implementation derives the advanced-capability rows from registered `Tool` metadata, keeps the discoverability tool in `core`, reports the active toolset from startup-provided registry options, rejects model-supplied arguments, and includes the `ICUVISOR_TOOLSET=full` restart instruction plus delete-mode guidance.

## Verification

- `git diff 5c378a0989eacb00845aef08eab1c2d1e74c2fe5..HEAD --name-only` — reviewed
- `git diff 5c378a0989eacb00845aef08eab1c2d1e74c2fe5..HEAD` — reviewed
- `go test ./...` — pass
- `make lint` — pass
