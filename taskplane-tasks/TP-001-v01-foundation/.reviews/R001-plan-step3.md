# Plan Review: TP-001 Step 3

Verdict: **Approved**

The updated Step 3 plan in `STATUS.md` now captures the concrete v0.1 config contract well enough to implement. It names the JSON fields and environment variables, defines the `--config` surface, documents precedence, defaults, validation behavior, app integration, redaction expectations, and narrows `.env` support to read-only recognized keys. This addresses the blocking gaps from the previous review.

## Notes to carry into implementation

1. **Make config-path precedence explicit in code/tests.** The plan implies `--config` wins over `ICUVISOR_CONFIG`; cover that in app/config tests so the CLI flag behavior is not ambiguous.

2. **Track presence separately from defaults.** The written precedence is `defaults < JSON < .env absent-only < process env < CLI flags`. Because `timezone`, `api_base_url`, and `http_timeout` have defaults, implement this with raw optional values or source tracking so `.env` can override built-in defaults while still not overriding JSON/process env values.

3. **Keep `version` config-free.** Add/adjust an app test proving `icuvisor version` does not attempt config loading or parse `--config`-related state.

4. **Reject malformed athlete IDs centrally.** The implementation should make the normalization contract precise: accept `12345` and `i12345`, emit `i12345`, and return a short actionable error for empty or malformed values.

5. **Preserve redaction beyond `String()`.** Ensure any validation/load errors that mention config sources or fields cannot include raw API key values. Prefer a redacted summary type for future logs.

6. **Constrain `.env` lookup.** If using automatic `.env` loading, keep it to the current working directory or another explicitly documented location; do not walk parent directories or print discovered values.

No further plan changes are required before coding Step 3.
