# Plan review — TP-053 Step 1

Decision: **changes requested** before executing the step.

The Step 1 checklist in `STATUS.md` is directionally right, but it is too loose for the highest-risk part of this tutorial: proving that ChatGPT can actually start/use icuvisor on a fresh macOS setup.

## Findings

1. **Blocking — the plan weakens “real machine / clean account” into “closest available macOS path”.**  
   The prompt requires running the tutorial flow on a clean macOS account or fresh user profile. `STATUS.md` currently says “closest available macOS path”, which could hide existing Keychain entries, config files, quarantine state, ChatGPT connector settings, or shell build artifacts. Update the plan to name the exact clean environment and cleanup steps, for example: macOS version, new user account, no existing `icuvisor` Keychain item, no prior `~/.config`/app config, no existing ChatGPT custom connector. If a truly clean profile is not available, record that as a blocker/limitation and get explicit approval rather than silently substituting.

2. **Blocking — ChatGPT MCP support needs a go/no-go validation, not just a doc link.**  
   The tutorial depends on the current ChatGPT custom-connector UX accepting the proposed local stdio JSON. Step 1 should explicitly verify and record:
   - canonical ChatGPT MCP/custom connector documentation URL and date accessed;
   - whether the required UI is in ChatGPT web, ChatGPT desktop, or both;
   - required plan/beta/enterprise flags, if any;
   - whether `transport: "stdio"` with `command` is accepted exactly as written;
   - the actual success/error state after connecting.

   If ChatGPT only supports remote HTTP connectors in the current UX, or if local stdio is gated in a way a typical beta athlete will not have, stop and escalate before drafting the tutorial.

3. **Major — the install path must be the same path the tutorial will teach.**  
   The prompt says to execute every step exactly as written. The plan should decide before testing whether Step 1 uses the signed DMG path or the build-from-source fallback, then time that exact path. Do not time a developer build while drafting DMG instructions, or vice versa. Also reconcile Gatekeeper handling with `docs/install/macos.md`, which says a notarized release should not require overriding Gatekeeper; if Gatekeeper appears, record whether the tutorial should verify signature, link to troubleshooting, or use the fallback path.

4. **Major — setup/keychain verification needs explicit secret-safety checks.**  
   Add checks that `icuvisor setup` is run without passing the API key through ChatGPT JSON, shell history, screenshots, or config files; that it stores the key in macOS Keychain; and that any Keychain permission prompt is captured in `STATUS.md`. This is central to the project’s hard rule that API keys never leave server config / OS keychain.

5. **Minor — timing data should be structured enough to drive simplification.**  
   The plan should require recording per-step timings for install, API key retrieval, setup, connector creation, and first question, plus a concrete action for every substep over 2 minutes. A single total time is not enough to identify where the tutorial is slow.

## Suggested revised Step 1 checklist

- Create or identify a clean macOS user/profile; record macOS version, ChatGPT surface/version, account capability, and cleanup performed.
- Verify current ChatGPT MCP custom-connector docs and UI; record canonical link, date accessed, transport support, required account flags, and exact accepted connector JSON.
- Choose the tutorial install path based on available release artifacts: signed DMG or build-from-source fallback. Test only that path.
- Run the full flow once from scratch: install, get intervals.icu API key, run `icuvisor setup`, connect ChatGPT, ask the first prompt.
- Record per-step timings and every unexpected Gatekeeper, Keychain, ChatGPT, or intervals.icu prompt.
- For each papercut, record the resolution that will appear in the tutorial or the reason it belongs in a footer/troubleshooting link instead.

