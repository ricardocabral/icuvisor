# KR5 benchmark methodology and results

This document defines the repeatable benchmark for PRD KR5: token efficiency versus the Python reference MCP servers.

## Scope

The benchmark compares four catalog surfaces:

1. `icuvisor-core` — icuvisor with `ICUVISOR_TOOLSET=core`; this is the headline KR5 surface.
2. `icuvisor-full` — icuvisor with `ICUVISOR_TOOLSET=full`.
3. `hhopke-intervals-icu-mcp` — the default `hhopke/intervals-icu-mcp` tool surface.
4. `mvilanova-intervals-mcp-server` — the default `mvilanova/intervals-mcp-server` tool surface, measured only as a black box.

## Metrics

### Per-session tool-description tokens

For each server, the harness opens a fresh MCP session and calls `tools/list`. It builds a stable JSON array from every registered tool:

```json
[{ "name": "...", "description": "...", "inputSchema": {} }]
```

The array is sorted by tool name and serialized as canonical compact UTF-8 JSON. The reported token count is the sum of tokens in that canonical catalog payload.

The pinned tokenizer is `cl100k_base` via `tiktoken==0.12.0`. The tokenizer package is MIT-licensed and is used only by the benchmark script, not by the shipped icuvisor binary. The result file records the tokenizer name and package version. If `tiktoken` is unavailable, the harness exits with installation guidance unless explicitly run with `--allow-approx-tokenizer` for smoke testing; approximate-tokenizer output is not accepted for KR5 results.

### Median per-call response bytes

For each shared prompt scenario, the harness executes the pinned MCP tool-call plan for each server. In live mode, every `tools/call` MCP `result` object is serialized as canonical compact UTF-8 JSON and measured in bytes. In fixture mode, redacted call fixtures may carry `redaction_audit.raw_response_bytes` inside the redacted content; when present, the harness counts that audited raw byte value, removes only the audit metadata, validates `redaction_audit.redacted_response_bytes` against the committed redacted MCP result, and validates that redacted bytes are within ±1% of raw. Calls without an audit field, including explicit unavailable/error results, are measured from their canonical MCP result JSON. The median is recorded as a JSON number; even-sized call sets may produce `.5` byte medians.

Only response bytes are counted; benchmark-only redaction padding/audit wrappers, transport framing, logs, latency, and the user's final natural-language answer are excluded.

## Shared prompt set and call-plan rules

The prompt set is pinned in `scripts/benchmark/prompts/kr5_shared_prompts.json` with version `kr5-forum-prompts-v1`. It extends the TP-016 v0.2 read dogfood prompts and TP-029 v0.3 safety/write prompts into ten forum-shaped scenarios: recent training review, recovery, weekly planning, race taper, activity detail, intervals/splits, wellness scales, calendar/workout-library, non-destructive note/message, and safe destructive-tool refusal.

Prompt text is vendor-neutral. It does not mention icuvisor, resources, toolset tiers, or server-specific tool names. Server-specific tool mappings live in the benchmark fixture/live configuration, not in the prompt text.

To avoid cherry-picking, call plans are fixed before measurement and follow these rules:

- Each prompt declares one or more abstract `required_intents` such as `recent_activities`, `fitness_trend`, or `wellness_window`.
- Every server maps each required intent to the minimum documented/default MCP tool call needed to answer that intent on that server.
- If a server requires multiple calls to satisfy one intent, all calls count toward the response-byte median.
- If a server lacks an equivalent tool, the harness records an explicit unavailable/error call result rather than dropping the prompt.
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

## Running

Fixture-mode reproducibility command for the committed result:

```bash
python3 scripts/benchmark/kr5_benchmark.py \
  --mode fixtures \
  --prompt-set scripts/benchmark/prompts/kr5_shared_prompts.json \
  --fixture-dir scripts/benchmark/testdata/fixtures \
  --output scripts/benchmark/results/kr5-results.json \
  --generated-at 2026-05-14T20:00:00Z
```

Live mode uses the same harness with `--mode live --config <config.json>`. Start from `scripts/benchmark/benchmark-config.example.json`, provide commands and environment outside the repository, and never commit secrets. When a live server lacks a required intent, set that call's `tool` to `unavailable:<intent>` in the private config so the harness records an explicit `isError=true` unavailable result instead of attempting `tools/call`.

## Current results

Committed fixture result: `scripts/benchmark/results/kr5-results.json`, generated from `kr5-forum-prompts-v1` at `2026-05-14T20:00:00Z`.

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

## KR5 verdict

KR5 is **partially confirmed** on the committed frozen snapshot. The response-byte target is confirmed against both Python references. The description-token target is not confirmed: `icuvisor-core` is 4,396 tokens versus a ≤4,338-token threshold for a 60% reduction against hhopke's 10,845-token surface, a gap of 58 tokens or 0.53 percentage points.

Gap `TP-034-KR5-DESC-001`: trim at least 60 `icuvisor-core` catalog tokens without weakening tool-selection clarity, then rerun this benchmark. If follow-up measurement shows those 60 tokens cannot be removed without materially degrading tool choice, propose recalibrating KR5's description-token target from ≥60% to ≥59% while keeping the ≥40% response-byte target unchanged.
