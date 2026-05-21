# R007 Plan Review — Step 3: Integrate with releases

Verdict: APPROVE

The Step 3 plan is appropriately scoped and matches the task requirements: it limits the first published MCPB artifact to the macOS universal slice, uses the existing release workflow rather than rebuilding a development binary, validates the manifest, signs/verifies the binary before packaging, uploads the `.mcpb` to the draft release, and includes it in final checksum regeneration.

## Notes to carry into implementation

- Prefer a pinned MCPB CLI version in release automation instead of `@latest`, or centralize the package version in an environment variable, so tagged releases are reproducible and not broken by an upstream CLI change.
- When selecting the GoReleaser universal binary, reuse the existing robust `find dist -path '*_darwin_all/icuvisor'` pattern or make the path explicit from GoReleaser metadata; do not assume a brittle exact directory name.
- The workflow should validate both the checked-in manifest during preflight and the staged/release manifest produced by `scripts/package_mcpb.sh`.
- The implementation should add the `.mcpb` pattern to the final artifact download and `sha256sum` command, otherwise the uploaded artifact will not be covered by the final `SHA256SUMS.txt`.
- After signing the standalone binary, run a verification command such as `codesign --verify --strict --verbose=2 <binary>` before invoking `scripts/package_mcpb.sh`, and package exactly that verified file.

No blocking plan changes are required before implementation.
