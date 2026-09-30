# Repository guidance

## Required checks

- Run focused package tests during development.
- Both `make test` and `make lint` must pass before completing API/CLI changes. Fix failures before marking the work complete.

## Architecture

- Wrap external dependencies behind interfaces as much as possible.
- All command subfolders use `shared.Client`, defined in `cli/cmd/shared/client.go` and exposed as `cmd.Client` in `cli/cmd/client.go`. Do not add command-specific client interfaces; the shared package avoids import cycles.
- Keep registration focused on Cobra metadata and flags, `RunE` on client/options wiring, action methods on validation and orchestration, and API wrappers in `api/`.
- Use generated OpenAPI clients through `api.Client` wrappers and format API errors with `api/errors.FormatAPIError`.
- Wrap errors with useful context and `%w`.

## CLI conventions

- Generally use verb-first commands, such as `list organization-members`, `add organization-member`, `delete organization-member`, and `update organization-member`.
- Command families live under `cli/cmd/<verb>/`, with registration, leaf commands, and tests colocated.
- Use `shared.RootOptions` for global options and injectable `ClientFactory` functions for command tests. Register commands with `shared.AddCmd` to inherit argument validation.
- Organization-scoped commands use `--org`/`-g` or `CS_ORG_ID`; `-O` remains a compatibility alias. Require a scope for member operations. Organization roles are strings (`admin`, `member`), while the team API uses numeric roles (the CLI accepts text roles).
- Member commands use `shared.ValidateUserID`: the public API accepts non-negative integer user IDs, including zero. Existence and permissions are checked by the API.
- List commands support table, JSON, and YAML output through `pkg/io`. Keep presentation in the command layer.
- Include the existing copyright/SPDX header in Go files and format changes with `gofmt`.

## Testing

- Always use Ginkgo and Gomega. Use external test packages and the existing suite setup; `go test` runs these suites.
- Write individual Ginkgo `It` cases. Do not use table-driven tests (`DescribeTable`, `Entry`, or loops over test cases).
- Use Mockery-generated mocks, not hand-written stubs or fakes.
- Avoid live API mutations in unit tests.

## Generated files

- After changing interfaces, regenerate mocks with `go tool mockery`, configured in `.mockery.yml`.
- Do not manually edit generated OpenAPI code or mocks.
- Docs and licenses are auto-generated on PRs; no need to generate them locally.

## Task-specific guidance

- When adding or updating CLI commands, read [the create-new-command skill](.agents/skills/create-new-command/SKILL.md).
- Keep detailed command-building procedures in that skill. The repository conventions above apply, including shared client interfaces and PR-generated docs/licenses.
