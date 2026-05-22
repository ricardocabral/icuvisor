---
title: "Terse by default"
description: "Why icuvisor keeps tool responses small unless you ask for full detail."
---

AI assistants have limited context windows. If every tool call returns raw streams, null-heavy wellness rows, and debug metadata, the assistant spends its attention on plumbing instead of coaching decisions.

icuvisor is terse by default so routine questions fit comfortably in a conversation:

- Read tools return the smallest useful shape first.
- Null values are stripped from terse responses, while meaningful zero, empty string, and `false` values remain.
- Response metadata explains scales and units when the assistant needs them.
- Small scalar custom fields that are useful in routine analysis stay in terse mode when they are safe to name directly. For example, `get_activity_intervals` returns manually-entered interval fields such as lactate under each row's `custom_fields` map when intervals.icu includes them.
- Heavy payloads, such as raw samples and full raw upstream objects, require an explicit `include_full: true` argument on tools that support it.

## What `include_full` means

`include_full` is an opt-in for detail. Use it when you are debugging, when the user explicitly asks for raw fields, or when the terse response is missing evidence needed for the answer.

It is not a "better answer" switch. Full payloads can be much larger and can make weaker or smaller-context models more likely to lose the thread.

## Toolset tiers are part of the same idea

icuvisor also keeps the default tool catalog small. The `core` tier covers daily-use tools, while `full` adds advanced or heavier workflows. This helps the AI client pick the right tool and reduces per-conversation tool-description load.

The exact `ICUVISOR_TOOLSET` values live in [Safety modes and toolset tiers]({{< relref "../reference/safety-modes#toolset-tier" >}}). The [`icuvisor_list_advanced_capabilities`]({{< relref "../reference/tools#icuvisor_list_advanced_capabilities" >}}) tool lets an assistant discover hidden full-tier tools and explain how to enable them.

## The goal

Terse-by-default does not hide data from you. It makes the common path fast, predictable, and token-efficient, while still allowing full detail when a prompt genuinely needs it.
