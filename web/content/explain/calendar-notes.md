---
title: "Calendar notes and event categories"
description: "Why there is no add_note tool — notes are just NOTE-category calendar events."
---

A common first guess is that icuvisor has an `add_note` tool. It does not, and the reason is worth understanding: on intervals.icu a note is not a separate kind of object, it is a *category* of calendar event.

## One event writer, many categories

Everything on the intervals.icu calendar is an event. A planned workout, a goal race, and a free-text note are all events that differ only by their `category`. icuvisor mirrors that model: a single writer, [`add_or_update_event`]({{< relref "/reference/tools#add_or_update_event" >}}), creates and updates all of them. To create a note, call it with `category: "NOTE"`.

## How a NOTE event is shaped

A NOTE event uses two fields:

- `name` — a short title. Give every new NOTE a non-empty one.
- `description` — the note body, as plain text. On updates, a supplied `description` replaces the NOTE body; omit it to leave the body unchanged.

Do not put the body in `workout_doc`. That field holds the structured-workout DSL and only makes sense for `WORKOUT` events; on a NOTE it is meaningless. See [Build and schedule workouts]({{< relref "/cookbook/build-workouts" >}}) for where `workout_doc` *does* belong.

A note can be as short as a reminder or as long as a plan:

```json
{
  "date": "2026-06-17",
  "category": "NOTE",
  "name": "Daily reminder",
  "description": "Take resting HR after waking, do 10 minutes mobility, and log sleep quality before training."
}
```

```json
{
  "date": "2026-06-12",
  "category": "NOTE",
  "name": "Race-week nutrition plan",
  "description": "Breakfast: oatmeal and banana. Lunch: rice bowl. Carry 90 g carbs/hour for the long ride."
}
```

The same shape covers travel logistics, coach annotations, or any other free-text entry you want pinned to a date. Like every write, creating a NOTE event needs the server in write mode — see [Why safety modes exist]({{< relref "safety-modes" >}}).
