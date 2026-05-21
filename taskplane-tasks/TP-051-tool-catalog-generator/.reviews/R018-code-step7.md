# R018 Code Review — Step 7: Replace duplicate prose

Decision: REQUEST CHANGES

I reviewed `git diff d306812..HEAD`, read the changed files, and grepped for remaining hardcoded catalog prose/counts in `README.md` and `web/layouts/`.

## Blocking finding

1. **`README.md` still contains a stale hardcoded tool-count claim.**  
   `README.md:24` still says `~25 MCP tools covering activities, wellness, fitness, events, and custom items.` Step 7 replaces the long hand-maintained list under `## MCP tool catalog`, but this remaining count/prose is another manually maintained catalog statement in the same README. It conflicts with TP-051's goal that the public catalog is generated from the registry, and it will drift just like the removed landing-page `~25 MCP tools` claim.

   Please remove the numeric claim or rephrase it without a fixed count, for example:
   - `Generated MCP tool catalog covering activities, wellness, fitness, events, and custom items.`
   - or point this feature bullet to the generated tool reference.

## Notes

- The `web/layouts/index.html` change correctly removes the stale `~25 MCP tools` landing-page claim and keeps the generated chip partial.
- The `web/README.md` note about `web/data/tools.json` being generated is appropriate.
- I did not run the full test suite; this review was limited to the Step 7 documentation/prose changes.
