# TP-053 — Getting-started tutorial (ChatGPT)

## Mission

Write a single Diataxis **tutorial** that takes a new icuvisor user from zero to having a working ChatGPT conversation about their intervals.icu training data. This is the only intentionally **learning-oriented** page on the website — the rest of the site is how-to, reference, and explanation. The user explicitly chose ChatGPT for the first tutorial (over Claude Desktop) because it is the most widely adopted MCP client among non-technical athletes.

Per Diataxis tutorial principles (loaded from the diataxis-documentation skill):

- Learning-oriented, not goal-oriented. The reader is **learning by doing**, not solving a real problem yet.
- Concrete and complete. Every command works on a fresh machine.
- Minimum viable confidence. The reader ends with one obviously-correct conversation, not a tour of the whole catalog.
- No branching ("if you're on Linux, …"). Pick one platform (macOS) and one client (ChatGPT). Other platforms/clients live as how-to guides (TP-052).
- No conceptual digressions. Link to explanations; do not inline them.

PRD anchors: §6 KR1 (install success — a clean tutorial is the single biggest lever).

ROADMAP positioning: v0.5 internal beta is the first audience; the tutorial must work end-to-end before beta athletes are onboarded.

Complexity: Blast radius 1 (one doc page), Pattern novelty 2 (no prior tutorial), Security 1, Reversibility 1 = 5 → Review Level 1. Size: S.

## Dependencies

- **TP-050** — Hextra scaffold with `tutorials/` section.
- **TP-052** — the how-to guides this tutorial links to (api-key, troubleshooting, after-upgrade). Tutorial does not duplicate them; it links once at the end.
- **TP-037** — macOS signed installer. If installer is not yet available, the tutorial uses the build-from-source path and notes "Until the v1.0 installer ships, build the binary first."

## Context to Read First

- `CLAUDE.md`
- `.claude/skills/diataxis-documentation/references/tutorials.md` (the canonical tutorial-writing guide)
- `docs/install/macos.md` (the source for the install steps)
- `docs/clients/claude-desktop.md` (template for client JSON snippet — adapt for ChatGPT)
- `README.md` §Quickstart
- ChatGPT MCP custom-connector docs — record the canonical link in `STATUS.md`. As of writing, ChatGPT supports MCP via custom connectors with stdio or HTTP transports; verify the current UX before writing screenshots.

## File Scope

Expected files (new):

- `web/content/tutorials/getting-started-chatgpt.md` — the tutorial page itself.
- `web/static/img/tutorials/chatgpt/` — screenshots (one per major step; cropped tight; PNG; no PII; alt text written for screen readers).
- `taskplane-tasks/TP-053-getting-started-tutorial-chatgpt/STATUS.md`.

Out of scope:

- Tutorials for Claude Desktop, Claude Code, Cursor, Continue, Zed, Pi — file follow-ups if user demand emerges.
- Conceptual content about MCP, local-first design, terse-by-default — those are explanation pages (TP-052).
- The reference material the tutorial links to.

## Tutorial structure

Write it in this order. Reader takes ~10 minutes from start to "first useful answer".

1. **What you'll do** (one short paragraph). "By the end you'll have asked ChatGPT about a real activity from your intervals.icu account and gotten a sourced answer."
2. **What you'll need.** Bulleted prerequisites: macOS 13+, a ChatGPT account with MCP custom connectors enabled, an intervals.icu account, ~10 minutes.
3. **Step 1 — Install icuvisor.** Cleanest path: DMG download → drag to Applications → `xattr -dr com.apple.quarantine /Applications/icuvisor.app` if Gatekeeper complains. If the installer is not yet shipped, fall back to `git clone … && make build` with a single line explaining why. One screenshot showing icuvisor in Applications.
4. **Step 2 — Get your intervals.icu API key.** Click-through from <https://intervals.icu/settings>. One screenshot of the "API Key" section, key value blurred.
5. **Step 3 — Run `icuvisor setup`.** Paste the key into the masked prompt. Setup verifies the key and stores it in the macOS keychain. No mention of legacy `.env` or `INTERVALS_ICU_API_KEY` — that is a how-to detail.
6. **Step 4 — Connect ChatGPT.** Open ChatGPT → Settings → Connectors → Add custom MCP → paste the JSON snippet. Provide the exact JSON the reader pastes:
   ```json
   {
     "name": "icuvisor",
     "command": "/Applications/icuvisor.app/Contents/MacOS/icuvisor",
     "transport": "stdio"
   }
   ```
   Two screenshots: the connector settings page; the "Connected" success state.
7. **Step 5 — Ask your first question.** Suggest a specific prompt: "Summarize my training load over the last 14 days using my intervals.icu data." Show a representative answer. Highlight that the answer cites tool calls (CTL/ATL/TSB) and links back to the user's data.
8. **What just happened.** One short paragraph: "ChatGPT asked icuvisor for your data through MCP. icuvisor talked to intervals.icu using your API key — which never left your Mac." Link out to `/explain/what-is-mcp/` and `/explain/local-first/`.
9. **Where to next.** Three links: (a) "Configure another AI client" → `/connect/`, (b) "Explore the full tool catalog" → `/reference/tools/`, (c) "Set up coach mode for a roster" → `/guides/coach-mode/`. **Do not** link to troubleshooting from inside the tutorial body — Diataxis says a tutorial should not anticipate failure; link in a footer instead.

## Steps

### Step 1: Verify the user flow on a real machine

- [ ] On a clean macOS account (or fresh user profile), execute every step exactly as written. Time it. If any step takes >2 minutes, simplify it.
- [ ] Note any step where Gatekeeper, ChatGPT UI, or keychain UI prompts the user unexpectedly. Either fold the resolution into the tutorial or link to troubleshooting.

### Step 2: Draft the page

- [ ] Write in second person ("you'll", "click", "paste"), present tense, no marketing voice.
- [ ] One concept per sentence. Short paragraphs.
- [ ] No "Note:" / "Tip:" interruptions. If something must be said, integrate it into the flow.
- [ ] Code blocks: full, copy-pasteable, no `<...>` placeholders the reader must mentally substitute. Use a fenced `text` block with the exact JSON; do not use `{{ }}` Hugo shortcodes inside the code.

### Step 3: Screenshots

- [ ] Capture at 2× retina, crop tight, redact any PII (API key, athlete ID, athlete name).
- [ ] Alt text: describe what changed compared to the previous step, not the chrome.
- [ ] Store under `web/static/img/tutorials/chatgpt/` with kebab-case filenames numbered by step (`01-applications.png`, `02-api-key.png`, …).
- [ ] Reference via `![alt](/img/tutorials/chatgpt/01-applications.png)`.

### Step 4: Build + verify

- [ ] `cd web && hugo --minify --gc` clean.
- [ ] Tutorial appears under "Tutorials" in nav.
- [ ] Pagefind indexes the page.
- [ ] Read the tutorial top-to-bottom on a phone-width viewport — verify screenshots are legible at that size.

### Step 5: External review

- [ ] Ask one non-technical user (the closer to the target audience the better) to follow the tutorial cold on a fresh machine. Note every place they hesitated. Fix.
- [ ] Record the reviewer's name and the date in `STATUS.md`.

## Acceptance Criteria

- A reader on macOS who has never seen icuvisor before can finish the tutorial in ~10 minutes and end up with a working ChatGPT conversation about their intervals.icu data.
- The tutorial does not duplicate content from how-to / reference / explanation pages — it links to them.
- Every step is a concrete action; no conditional branches; no "if you're on Linux".
- Every screenshot has PII-free content and accessible alt text.
- The tutorial has been completed cold by at least one person who is not the author.

## Do NOT

- Do not turn the tutorial into a tour of the full tool catalog — one good example, not ten.
- Do not branch by OS or client inside the tutorial body — pick one path.
- Do not explain MCP, local-first, terse-by-default in the tutorial body. Link to explanation pages.
- Do not anticipate failure paths inside the body. Troubleshooting lives elsewhere; mention it only in a footer.
- Do not add emojis (per CLAUDE.md).
- Do not include marketing claims ("the best", "the fastest"); the tutorial sells competence, not the product.
- Do not hardcode an example athlete ID, API key, or activity ID in the JSON snippet beyond what the user pastes.

## Documentation

Must update:

- `STATUS.md` — reviewer name + date, timing of the cold-run, any UX papercuts surfaced and how they were resolved.
- `CHANGELOG.md` `[Unreleased]` — "Added" entry referencing the new tutorial.

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-053`, e.g., `TP-053 draft chatgpt getting-started tutorial`.

---

## Amendments

_Add amendments below this line only._
