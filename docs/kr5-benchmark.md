# KR5 benchmark methodology and results

This document defines the repeatable benchmark for PRD KR5: token efficiency versus the Python reference MCP servers.

## Scope

The benchmark compares four historical KR5 catalog surfaces plus one scoped analyzer-family fixture:

1. `icuvisor-core` — icuvisor with `ICUVISOR_TOOLSET=core`; this is the headline KR5 surface.
2. `icuvisor-full` — icuvisor with `ICUVISOR_TOOLSET=full`.
3. `hhopke-intervals-icu-mcp` — the default `hhopke/intervals-icu-mcp` tool surface.
4. `mvilanova-intervals-mcp-server` — the default `mvilanova/intervals-mcp-server` tool surface, measured only as a black box.
5. `icuvisor-analyzer-family` — deterministic synthetic fixture for analyzer-enabled versus analyzer-disabled call plans. Its `analyzers_enabled` and `analyzers_disabled` modes share byte-identical non-analyzer catalog payloads; enabled mode adds the analyzer family only.

## Metrics

### Per-session tool-description tokens

For each server, the harness opens a fresh MCP session and calls `tools/list`. It builds a stable JSON array from every registered tool:

```json
[{ "name": "...", "description": "...", "inputSchema": {} }]
```

The array is sorted by tool name and serialized as canonical compact UTF-8 JSON. The reported token count is the sum of tokens in that canonical catalog payload.

The pinned tokenizer is `cl100k_base` via `tiktoken==0.12.0`. The tokenizer package is MIT-licensed and is used only by the benchmark script, not by the shipped icuvisor binary. The result file records the tokenizer name and package version. If `tiktoken` is unavailable, the harness exits with installation guidance unless explicitly run with `--allow-approx-tokenizer` for smoke testing; approximate-tokenizer output is not accepted for KR5 results.

### Per-call response bytes, response tokens, and raw-stream pulls

For each shared prompt scenario, the harness executes the pinned MCP tool-call plan for each server. In live mode, every `tools/call` MCP `result` object is serialized as canonical compact UTF-8 JSON and measured in bytes and tokens. In fixture mode, redacted call fixtures may carry `redaction_audit.raw_response_bytes` inside the redacted content; when present, the harness counts that audited raw byte value, removes only the audit metadata, validates `redaction_audit.redacted_response_bytes` against the committed redacted MCP result, and validates that redacted bytes are within ±1% of raw. Response-token metrics strip benchmark-only `redaction_audit` metadata before tokenization. Calls without an audit field, including explicit unavailable/error results, are measured from their canonical MCP result JSON. Medians are recorded as JSON numbers; even-sized call sets may produce `.5` medians.

Only response payloads are counted; benchmark-only redaction padding/audit wrappers, transport framing, logs, latency, and the user's final natural-language answer are excluded.

For analyzer modes, the harness also records LLM-visible `raw_stream_pull_count`: the count of call-plan rows whose top-level tool is `get_activity_streams` or a configured reference alias. Internal analyzer `_meta.source_tools` entries remain separate evidence and do not count as LLM-visible raw-stream pulls.

## Shared prompt set and call-plan rules

The prompt set is pinned in `scripts/benchmark/prompts/kr5_shared_prompts.json` with version `kr5-forum-prompts-v1`. It extends the TP-016 v0.2 read dogfood prompts and TP-029 v0.3 safety/write prompts into ten forum-shaped historical scenarios, plus five analyzer-family scenarios scoped to `icuvisor-analyzer-family`: trend, zone-time distribution, baseline, correlation/compliance, and single-activity histogram.

Prompt text is vendor-neutral. It does not mention icuvisor, resources, toolset tiers, or server-specific tool names. Server-specific tool mappings live in the benchmark fixture/live configuration, not in the prompt text.

To avoid cherry-picking, call plans are fixed before measurement and follow these rules:

- Each prompt declares one or more abstract `required_intents` such as `recent_activities`, `fitness_trend`, or `wellness_window`.
- Prompts may declare `server_scope`; scoped analyzer prompts apply only to the analyzer fixture and are not missing coverage for legacy/reference fixtures.
- Every applicable server maps each required intent to the minimum documented/default MCP tool call needed to answer that intent on that server.
- If a server requires multiple calls to satisfy one intent, all calls count toward the response-byte median.
- If a server lacks an equivalent tool, the harness records an explicit unavailable/error call result rather than dropping the prompt.
- Analyzer prompts declare expected top-level tools and expected source-tool usage by mode. The enabled and disabled modes use the same prompt text; only tool availability and fixed call-plan rows differ.
- Prompt text and intent lists are reviewed before any server metrics are inspected.

## Frozen account snapshot

The committed results use snapshot `kr5-redacted-test-athlete-v1`, described by `scripts/benchmark/testdata/snapshot-manifest.json`.

Snapshot contents:

- Date windows: shifted last-14-days activities, last-42-days fitness, last-21-days wellness, next-14-days calendar events, and one shifted race-week window.
- Entities: one redacted athlete profile, representative ride/run activities, one activity with intervals/splits/messages, one wellness row containing both `sleepQuality` and `sleepScore`, upcoming events/training-plan/workout-library summaries, and synthetic non-destructive write confirmations.
- Reference surfaces: black-box `tools/list` outputs and redacted `tools/call` result shapes for the two Python references. The `mvilanova` fixture is derived only from running the server and capturing JSON-RPC output; no GPL source was read or copied.
- Redaction: athlete IDs, activity/event/workout IDs, comments, names, exact dates, threshold values, body metrics, and account-specific text are replaced with placeholders or shifted synthetic values.
- Byte policy: redaction is performed before committing; live raw response byte counts are preserved as `redaction_audit.raw_response_bytes`, with `redacted_response_bytes` validated against the committed redacted MCP result after removing audit metadata and required to remain within ±1% per audited call. Audit fields are not counted as response bytes.

## Non-determinism policy

The committed benchmark runs in fixture mode by default so CI and contributors can reproduce results without a real intervals.icu API key or access to the reference servers. Fixture mode is the authoritative reproducibility gate for this repository.

Live mode is supported for recalibration. A live run must use the same prompt-set version, the same dedicated test athlete account snapshot manifest, and exact server versions recorded in the result file. Raw live transcripts must not be committed; only redacted fixture/result files are allowed.

Acceptable fixture rerun tolerance is zero for catalog token counts and zero for response-byte medians because both are computed deterministically from committed canonical JSON and audited byte fields. Live reruns should stay within ±5% response-byte median unless upstream data changed; outside that range, refresh the frozen snapshot and document why.

## First-tool routing smoke eval

The focused routing eval in `internal/toolrouting` and `scripts/toolroutingeval` covers first-tool-call selection for ambiguous prompts. It is separate from KR5: KR5 measures fixed call-plan token and response efficiency, while the routing eval checks whether catalog descriptions make the first MCP tool choice measurable for prompts such as arbitrary historical activity-window lists that should start with `get_activities`, ATP/periodization summaries versus raw calendar events, training-plan assignment, and fitness projection.

Default execution is credential-free and CI-safe:

```bash
make eval-tool-routing
```

With no provider configured, the runner validates the committed fixture (`internal/toolrouting/testdata/cases.json`), loads the registered tool catalogs for each `catalog_mode`, confirms expected tools are exposed, reports every case as skipped, and exits successfully. It does not call intervals.icu and does not make an LLM request.

Saved-result and diff workflows are available for explicit local runs:

```bash
go run ./scripts/toolroutingeval -json -output /tmp/routing-current.json
go run ./scripts/toolroutingeval -diff /tmp/routing-baseline.json -json -output /tmp/routing-diff.json
```

Diff output classifies each case as `win`, `regression`, `still_failing`, `still_passing`, `new`, `removed`, or `skipped`. Human output never prints raw provider messages. JSON output omits raw provider messages by default; pass `-include-raw` only for local debugging and do not commit unsanitized model responses.

Optional model-backed routing is opt-in:

```bash
ICUVISOR_ROUTING_EVAL_PROVIDER=anthropic \
ANTHROPIC_API_KEY=... \
go run ./scripts/toolroutingeval -json -output /tmp/routing-live.json
```

The API key is read from the environment only. The default test suite and `make eval-tool-routing` remain skipped-by-default and require no paid model call.

## Running

Fixture-mode reproducibility command for the committed result:

```bash
python3 scripts/benchmark/kr5_benchmark.py \
  --mode fixtures \
  --prompt-set scripts/benchmark/prompts/kr5_shared_prompts.json \
  --fixture-dir scripts/benchmark/testdata/fixtures \
  --output scripts/benchmark/results/kr5-results.json \
  --generated-at 2026-05-20T00:00:00Z
```

Live mode uses the same harness with `--mode live --config <config.json>`. Start from `scripts/benchmark/benchmark-config.example.json`, provide commands and environment outside the repository, and never commit secrets. When a live server lacks a required intent, set that call's `tool` to `unavailable:<intent>` in the private config so the harness records an explicit `isError=true` unavailable result instead of attempting `tools/call`.

## Current results

Committed fixture result: `scripts/benchmark/results/kr5-results.json`, schema `kr5-benchmark-result-v2`, generated from `kr5-forum-prompts-v1` at `2026-05-20T00:00:00Z` with 15 prompts. The historical KR5 comparison below uses each server's `default` mode.

| Server                           | Version                                                                               | Tools | Description tokens | Median response bytes |
| -------------------------------- | ------------------------------------------------------------------------------------- | ----: | -----------------: | --------------------: |
| `icuvisor-core`                  | `cc566c3-dirty`                                                                       |    17 |              4,396 |                 976.5 |
| `icuvisor-full`                  | `cc566c3-dirty`                                                                       |    38 |              9,490 |               1,154.0 |
| `hhopke-intervals-icu-mcp`       | `intervals-icu-mcp==2.0.0`, tag `v2.0.0` (`d6d8f2b381db0776b0bb6d3ff1081d733bf0ac96`) |    58 |             10,845 |               2,063.5 |
| `mvilanova-intervals-mcp-server` | `0.1.0` at `12199c61d88f580a885f04921b23dcf7c4524de8`                                 |    17 |              6,227 |               1,649.5 |

Headline KR5 deltas use `icuvisor-core`:

| KR5 metric              | Baseline               | Target | Measured reduction | Result |
| ----------------------- | ---------------------- | -----: | -----------------: | ------ |
| Tool-description tokens | hhopke 58-tool surface |   ≥60% |             59.47% | Miss   |
| Median response bytes   | hhopke                 |   ≥40% |             52.68% | Pass   |
| Median response bytes   | mvilanova              |   ≥40% |             40.80% | Pass   |

## Analyzer-family comparison

The analyzer fixture is a deterministic call-plan benchmark, not an autonomous model tool-selection test. It compares identical analyzer prompt text across two modes: `analyzers_enabled` exposes the analyzer-family catalog, while `analyzers_disabled` exposes the same non-analyzer catalog and uses fetch-and-reduce call rows.

| Mode | Tools | Description tokens | Calls | Response tokens | Median response tokens | Median response bytes | Raw stream pulls |
| ---- | ----: | -----------------: | ----: | --------------: | ---------------------: | --------------------: | ---------------: |
| `analyzers_disabled` | 38 | 9,490 | 9 | 2,266 | 251 | 1,357 | 3 |
| `analyzers_enabled` | 49 | 12,063 | 6 | 321 | 53 | 216 | 0 |

Aggregate analyzer response-token reduction is 85.83% (`2,266 → 321`) and LLM-visible raw-stream pulls drop from 3 to 0. The enabled rows for the roadmap trend, zone-time distribution, and correlation/compliance shapes make zero top-level raw-stream calls; internal analyzer source-tool evidence is recorded separately in `source_tool_usage`.

### TP-098 core-promotion evidence

Per-candidate net savings are computed from paired prompt rows using response tokens, then subtracting that candidate tool's incremental catalog-description tokens measured with the same canonical catalog payload and tokenizer conventions. A candidate `meets` if net savings are positive and the enabled row has no LLM-visible raw-stream pull.

| Candidate | Prompt | Enabled response tokens | Disabled response tokens | Gross response-token savings | Incremental catalog tokens | Net savings | Raw stream pulls | TP-098 gate |
| --------- | ------ | ----------------------: | -----------------------: | ---------------------------: | -------------------------: | ----------: | ---------------: | ---------- |
| `analyze_trend` | `KR5-A01` | 48 | 502 | 454 | 305 | 149 | `1 → 0` | Meets |
| `compute_zone_time` | `KR5-A02` | 50 | 253 | 203 | 173 | 30 | `1 → 0` | Meets |
| `compute_baseline` | `KR5-A03` | 49 | 502 | 453 | 234 | 219 | `0 → 0` | Meets |

These results provide positive fixture evidence for TP-098's benchmark-gated core-promotion candidates. TP-100 does not promote the tools; it records evidence for TP-098 to decide placement.

## KR5 verdict

KR5 remains **partially confirmed** on the committed frozen snapshot. The response-byte target is confirmed against both Python references. The description-token target is not confirmed: `icuvisor-core` is 4,396 tokens versus a ≤4,338-token threshold for a 60% reduction against hhopke's 10,845-token surface, a gap of 58 tokens or 0.53 percentage points.

Gap `TP-034-KR5-DESC-001`: trim at least 60 `icuvisor-core` catalog tokens without weakening tool-selection clarity, then rerun this benchmark. If follow-up measurement shows those 60 tokens cannot be removed without materially degrading tool choice, propose recalibrating KR5's description-token target from ≥60% to ≥59% while keeping the ≥40% response-byte target unchanged.
