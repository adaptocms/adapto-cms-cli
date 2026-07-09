# Development

## Make targets

| Target | Effect |
|--------|--------|
| `make build` | Build `adapto`, version stamped `dev-<commit>` |
| `make test` | Build, vet, and run all tests with the race detector |
| `make check-docs` | Fail if the README command tree is stale |
| `make generate` | Fetch the OpenAPI spec and regenerate `internal/client` (override source with `SPEC_URL=...`) |
| `make docs` | Regenerate the README command tree |
| `make release` | Tag and push a release (`BUMP=minor` or `BUMP=major` for bigger bumps) |

## Tests

Unit tests live next to their packages in `internal/`. Integration tests (`test/integration`) build the real binary and run it against the mock Management API in `test/mockapi`. Mock fixtures use the generated client models, so API schema drift fails the compile after `make generate`.

## Generated docs

The README command tree and the `llm-info` reference are generated from the cobra command tree. Document commands in `Short`/`Long`/`Example` and flag usage strings, not in markdown. CI fails if the README tree is stale; run `make docs`.

## Branches and releases

- `develop`: integration branch, CI only.
- `main`: every push releases. CI tags the next patch version and publishes binaries for all platforms.
  - `[release minor]` or `[release major]` in the commit message changes the bump.
  - `[skip release]` skips the release.
- `make release` (manual tag push) also triggers the release build.
