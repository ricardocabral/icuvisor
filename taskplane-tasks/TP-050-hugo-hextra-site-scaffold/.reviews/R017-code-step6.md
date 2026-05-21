# Code Review R017 - Step 6: Local preview docs

Status: REVISE

## Findings

1. `web/README.md:28` documents the production-style Pagefind preview as `python3 -m http.server --directory public 1313`. Python's `http.server` binds to all interfaces by default, so this local preview command exposes the generated site on the LAN. The project guidance explicitly says local HTTP surfaces should not bind beyond `127.0.0.1` by default unless LAN exposure is opt-in and documented. Please change the command to bind loopback, for example:

   ```bash
   python3 -m http.server --bind 127.0.0.1 --directory public 1313
   ```

   The following line can continue to direct users to `http://localhost:1313`.

## Notes

- The Hugo Modules preview flow, Pagefind indexing command, Hextra pin documentation, and deploy summary otherwise match the Step 6 requirements.
