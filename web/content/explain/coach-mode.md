---
title: "Coach mode model"
description: "How icuvisor targets multiple athletes without turning athlete IDs into credentials."
---

Coach mode lets one local icuvisor process use a coach-scoped intervals.icu API key across a configured roster. The key stays in the server's credential chain; the AI client sees athlete selectors and allowed tools, not the key.

## The server holds the credential

The coach API key is loaded the same way as a single-athlete key: process environment, OS keychain, or legacy local config fallback. It is never accepted as a tool argument and never returned in tool output.

That matters because chat text is not a credential store. The model can ask to target an athlete, but it cannot receive or supply the upstream API key.

## `athlete_id` is a selector

In coach mode, `athlete_id` tells icuvisor which configured athlete to target. It is not proof of authorization.

Before calling intervals.icu, icuvisor normalizes the ID, checks that it is in `coach.athletes`, and evaluates the selected athlete's ACL. A malformed ID, an athlete outside the roster, or an ACL-denied target gets a short generic error so roster membership is not exposed through detailed error text.

## ACLs are another gate

Coach ACLs compose with the global gates:

1. Delete/write mode decides whether read, write, and delete tools can register.
2. Toolset tier decides whether core or full tools can register.
3. Per-athlete ACL decides what the active athlete can use.

Any deny wins. A coach can make an active client full-access except deletes, a prospect read-only, and a paused client deny-all.

## Selection and stale catalogs

[`list_athletes`]({{< relref "../reference/tools#list_athletes" >}}) returns the configured roster. [`select_athlete`]({{< relref "../reference/tools#select_athlete" >}}) changes the default target for subsequent calls in the same MCP session.

MCP clients may cache the tool catalog for a conversation. If switching athletes changes which tools are visible, [`select_athlete`]({{< relref "../reference/tools#select_athlete" >}}) can return `_meta.requires_new_conversation: true`. Start a new conversation or reconnect before relying on the changed catalog.

## Setup guide

For the JSON shape, mode values, and examples, see [Set up coach mode]({{< relref "../guides/coach-mode" >}}). For the broader privacy trust boundary around coach-scoped credentials and AI clients, see [Privacy posture]({{< relref "privacy" >}}).
