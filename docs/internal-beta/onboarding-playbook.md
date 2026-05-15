# v0.5 internal beta onboarding playbook

Use this after a participant passes [protocol](protocol.md) screening and consents. Do not collect raw training data, API keys, athlete IDs, tool arguments, or screenshots with values.

## Operator terminal recipe

1. Download the current signed macOS DMG from the maintainer-provided release candidate.
2. Install using the [macOS install guide](../install/macos.md); verify Gatekeeper opens the app/binary.
3. Open Terminal and confirm the installed binary:
   ```sh
   icuvisor version
   icuvisor --help
   ```
4. Run first-time setup with the participant typing their own API key:
   ```sh
   icuvisor setup
   ```
5. Have the participant configure their client using the canonical snippets, without copying JSON here:
   - [Claude Desktop setup](../clients/claude-desktop.md)
   - [Claude Code setup](../clients/claude-code.md)
6. Restart the client and start a fresh conversation.
7. First-call verification prompt:
   > Use icuvisor to summarize my athlete profile. Do not reveal raw IDs or API keys.
8. Record install-to-first-successful-tool-call time in [findings](findings.md).

## Coach-mode variant

Use only for coach participants who consented to redacted coach-mode feedback.

1. Review [coach mode](../coach-mode.md) and TP-039 roster/ACL behavior with the coach.
2. Explain that `athlete_id` is a selector, not a credential, and should not be sent in feedback.
3. Help them create a local roster config with labels they understand; do not ask them to share roster names, athlete IDs, or athlete data.
4. Verify only redacted categories:
   - Can list configured athletes.
   - Can select a target athlete.
   - ACLs hide denied tools.
   - Coach understands catalog-cache guidance after config changes.
5. Capture coach-mode blockers as categories only, for example `acl-confusing` or `selection-state unclear`.

## Troubleshooting

| Symptom | Operator action |
| --- | --- |
| Client cannot find server | Re-check the linked Claude Desktop/Code doc and restart the client. |
| Setup cannot load credentials | Re-run `icuvisor setup`; participant types secrets locally. |
| Tool catalog seems stale | Start a new client conversation; for coach-mode changes, follow catalog-cache guidance in `docs/coach-mode.md`. |
| Need local support bundle | Ask participant to run `icuvisor diagnostics` and voluntarily paste only that redacted output. |
| HTTP transport confusion | Use stdio for beta unless the participant is explicitly testing Streamable HTTP. |

## Mid-beta update exercise

If the maintainer ships a replacement release candidate during the beta, document this as an instruction only:

1. Download the new signed DMG.
2. Quit the MCP client.
3. Install over the prior build.
4. Run `icuvisor version` and `icuvisor diagnostics`.
5. Start a fresh conversation and repeat first-call verification.

Do not retag a release or perform the exercise as part of TP-041.
