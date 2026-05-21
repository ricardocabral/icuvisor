---
title: "Reference"
description: "Generated icuvisor reference pages."
weight: 50
type: docs
cascade:
  type: docs
---

Exact names, fields, flags, and values for the icuvisor binary and its MCP surface. For task-oriented steps, see the [Guides]({{< relref "/guides" >}}) and [Tutorials]({{< relref "/tutorials" >}}).

Several pages are generated from the repository source of truth (tool registry, CLI golden fixture), so they always match the shipped binary.

{{< cards >}}
  {{< card link="tools" title="Tool reference" subtitle="Every MCP tool icuvisor registers, by domain, with toolset tier and safety gate." >}}
  {{< card link="resources-prompts" title="MCP resources and prompts" subtitle="The MCP Resources and Prompts icuvisor exposes." >}}
  {{< card link="cli" title="CLI reference" subtitle="Commands, flags, environment variables, and exit codes." >}}
  {{< card link="config-file" title="Config file reference" subtitle="Every config-file JSON field, plus the default config path per platform." >}}
  {{< card link="safety-modes" title="Safety modes and toolset tiers" subtitle="How ICUVISOR_DELETE_MODE gates write/delete tools and how toolset tiers are selected." >}}
{{< /cards >}}
