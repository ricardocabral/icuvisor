# Plan review — TP-053 Step 3

Decision: **changes requested** before creating screenshots.

The Step 3 checklist has the right direction — create numbered PNGs, redact sensitive data, be honest about unavailable ChatGPT UI, and record the limitation — but it is still too underspecified for the known constraints from Steps 1 and 2. This step is where the tutorial can most easily become misleading or leak personal training data, so the plan needs a more concrete artifact list and capture/redaction rules before implementation.

## Findings

1. **Blocking — the plan does not name the exact six screenshot assets and their source.**  
   The draft page currently references:
   - `/img/tutorials/chatgpt/01-install.png`
   - `/img/tutorials/chatgpt/02-api-key.png`
   - `/img/tutorials/chatgpt/03-setup.png`
   - `/img/tutorials/chatgpt/04-connector.png`
   - `/img/tutorials/chatgpt/05-connected.png`
   - `/img/tutorials/chatgpt/06-first-answer.png`

   Step 3 should explicitly plan each file, whether it is a real capture, a redacted capture, or an illustrative/simulator image. This is especially important for the ChatGPT connector and connected-state screenshots because `STATUS.md` records that real ChatGPT UI was unavailable and current public docs do not match the local-stdio JSON flow. Do not leave the worker to improvise those images.

2. **Blocking — the plan must define how simulated ChatGPT images will be labeled honestly.**  
   The current tutorial alt text says things like “ChatGPT connector settings…” and “ChatGPT showing…” as if those are real ChatGPT captures. The Step 3 plan says to update alt text/captions, but it should require concrete wording such as “Illustration of the ChatGPT connector setup…” or “Simulator view of the connected state…”. The page must not fabricate evidence of a live ChatGPT UI state that was not available during verification.

3. **Major — the plan omits the prompt’s image-quality requirements.**  
   The task requires screenshots captured at 2× retina and cropped tight. Add those requirements to the Step 3 checklist, along with a quick verification pass that the images remain legible at phone width. The latter is formally in Step 4, but screenshot production should avoid creating assets that are obviously too wide or too blurry.

4. **Major — the redaction rule needs to cover aggregate training data in `06-first-answer.png`.**  
   `STATUS.md` says the Codex validation used real maintainer credentials and produced aggregate totals. The Step 3 checklist says to redact private training details, but it should explicitly forbid putting the real validation totals, dates, activity mix, athlete/account names, URLs, or tool-call payloads into the screenshot. Use synthetic illustrative values, blur the numbers, or crop to show only the existence of tool use and answer structure.

5. **Minor — record the limitation in `STATUS.md` with enough detail to audit later.**  
   The plan should say that `STATUS.md` will record which images are real captures versus illustrative/simulator assets, why ChatGPT UI images were not real captures, and what redaction/synthetic-data approach was used.

## Suggested revised Step 3 checklist

- Create `web/static/img/tutorials/chatgpt/`.
- Produce the six PNGs currently referenced by the tutorial, with this per-file plan recorded before capture:
  - `01-install.png`: real Terminal capture of the deterministic `/Users/Shared/icuvisor/bin/icuvisor version` flow, cropped tight, no username/worktree/private path beyond the tutorial path.
  - `02-api-key.png`: real or illustrative intervals.icu settings/API-key section with the key, athlete name, athlete ID, and account details fully redacted or synthetic.
  - `03-setup.png`: real or illustrative Terminal setup completion view with the API key masked and no actual athlete ID/name/timezone beyond harmless tutorial values.
  - `04-connector.png`: clearly labeled illustrative/simulator image for the local-stdio ChatGPT connector JSON unless a real ChatGPT local-stdio UI capture is available.
  - `05-connected.png`: clearly labeled illustrative/simulator connected-state image unless a real ChatGPT local-stdio connected-state capture is available.
  - `06-first-answer.png`: clearly labeled illustrative/simulator first-answer image with no real private training totals, tool payloads, names, URLs, activity titles, or locations.
- Capture or render at 2× retina, crop tightly, save as PNG, and keep filenames numbered/kebab-case matching the markdown references.
- Update the markdown alt text so each image describes what changed in the step and honestly identifies illustrative/simulator material where applicable.
- Verify every image manually for PII/secrets before committing: API key, athlete ID, athlete name, activity ID/title, locations, URLs, usernames, shell history, private paths, and real training details.
- Record in `STATUS.md` the screenshot limitation, which assets are illustrative/simulator, and the redaction/synthetic-data choices.

Once those specifics are added, the Step 3 plan should be ready to implement.
