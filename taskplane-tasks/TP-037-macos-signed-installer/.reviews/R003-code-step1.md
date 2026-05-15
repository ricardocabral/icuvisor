# Code review — TP-037 Step 1

Decision: **changes requested**

## Blocking findings

1. **Step 1 now appears complete even though the Developer ID gate is still unmet.**  
   `STATUS.md:17` replaces the original unchecked `Developer ID cert enrolled, .p12 exportable for CI` item with a checked documentation/deferment item, while `STATUS.md:72` still records `0 valid identities found` and missing Team ID/cert/common-name/GitHub-secret-presence inputs. That contradicts the task's Step 1 requirement to confirm the Developer ID certificate and the R002 carry-forward requirement not to check that item complete until the real non-secret facts are supplied. Even if steering accepted that the live certificate setup is maintainer-owned, the status should keep an explicit unchecked operator gate (or otherwise make Step 1 visibly blocked) so later Step 2/3 work cannot treat all Step 1 checklist items as done.

## Non-blocking notes

- `STATUS.md:88` contains a truncated note (`real non-secret me`) and the file currently lacks a trailing newline. Please clean this up while touching the status file.
- The new `SECURITY.md` preflight is a good place for the release gate and it correctly avoids committing secret material. Consider tightening `release notes or task status` to whichever location is authoritative for TP-037 so the required Team ID/common name/expiration are not lost between task execution and release.

## Validation

- Reviewed `git diff 2a3f05e..HEAD --name-only` and full diff.
- No code/tests to run for these documentation-only changes.
