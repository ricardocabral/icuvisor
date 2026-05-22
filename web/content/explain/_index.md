---
title: "Explanation"
description: "Concepts behind icuvisor's design."
weight: 60
type: docs
cascade:
  type: docs
---

Background on why icuvisor works the way it does — read these to build a mental model, not to complete a task.

{{< cards >}}
  {{< card link="what-is-mcp" title="What is MCP?" subtitle="A short explanation of Model Context Protocol for icuvisor users." >}}
  {{< card link="local-first" title="Local-first design" subtitle="Why icuvisor keeps credentials and data flow on your computer." >}}
  {{< card link="terse-by-default" title="Terse by default" subtitle="Why tool responses stay small unless you ask for full detail." >}}
  {{< card link="safety-modes" title="Why safety modes exist" subtitle="The reasoning behind safe, full, and none delete/write modes." >}}
  {{< card link="coach-mode" title="Coach mode model" subtitle="How icuvisor targets multiple athletes without making athlete IDs credentials." >}}
  {{< card link="fitness-projection" title="How fitness projection works" subtitle="Why get_fitness_projection is a deterministic scenario model, not a forecast." >}}
  {{< card link="interval-sources" title="Structured workouts vs. device laps" subtitle="How to read interval data without inventing workout steps that were never planned." >}}
  {{< card link="calendar-notes" title="Calendar notes and event categories" subtitle="Why there is no add_note tool — notes are NOTE-category calendar events." >}}
{{< /cards >}}
