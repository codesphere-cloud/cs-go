---
name: create-new-command
description: "Use when creating or updating a new CLI command in cli/cmd. Covers the folder-based command layout, responsibility boundaries, registration, output formatting, and the Ginkgo/Gomega test pattern used in this repo."
---

# Create a New Command

Use this skill when adding a new `cobra` command under `cli/cmd`. The preferred shape is now a command family folder: the parent command file and all direct subcommands live together in a subfolder, and the command logic stays split into registration, client wiring, business logic, and output formatting.

## Folder Layout

For a new command family, create a subfolder under `cli/cmd` named after the family.

- Put the parent command file in that folder, for example `cli/cmd/start/start.go`.
- Put all direct subcommands for that family in the same folder, for example `cli/cmd/start/pipeline.go`.
- Put the tests for that family in the same folder, for example `cli/cmd/start/pipeline_test.go`.
- Use deeper nested folders when a command family itself has sub-families, for example `cli/cmd/team/member/*.go`.
- Keep `cli/cmd` focused on shared command infrastructure, shared options, and the top-level root command wiring.
- Give each command family a small shared options/dependencies struct for values used by all children in that family.
- Keep child-specific flags and options inside the child command file instead of passing them through the parent command.
- For example, the `start` family should pass shared workspace/client dependencies through the parent, while `pipeline` owns its own stage flags and timeout settings.
- Prefer initializing that shared family state in the parent command's `PersistentPreRunE`.
- Let `AddStartCmd`-style entry points accept only the root command, then build the family state internally and hand it to children.

The existing organization list command still shows the best implementation pattern for a simple leaf command, especially the split between registration, client wiring, business logic, and output formatting.

Reference files:

- `cli/cmd/list_organizations.go` shows the full split between `Add...Cmd`, `RunE`, and the method that performs the actual work.
- `cli/cmd/list_organizations_test.go` shows the current external test style with `cmd_test`, `Ginkgo`, `Gomega`, `MockEnv`, `MockClient`, stdout capture, and Cobra registration checks.
- `cli/cmd/start.go` and `cli/cmd/start_pipeline.go` show the current parent/subcommand split that should be moved into a shared folder in the new layout.

## Function Responsibilities

Keep each function narrow.

### 1. `Add...Cmd`

This function should only build and register the Cobra command.

Typical responsibilities:

- Create the `cobra.Command` with `Use`, `Short`, `Long`, and `Example`.
- Store the command options struct on the command wrapper.
- Set `RunE` to the wrapper method.
- Attach the command to the parent with `AddCmd`.

Do not put API calls or output logic here.

### 2. `RunE`

This function should only coordinate execution.

Typical responsibilities:

- Create a client with `NewClient` or an injected `ClientFactory` when tests need control.
- Convert client creation errors into a command-level error message.
- Call the command’s action method.
- Convert action errors into a command-level error message.

Do not format tables or print JSON/YAML here.

### 3. The action method

This is where the command’s real work belongs.

For list commands, this method should:

- Call the client method that fetches data.
- Handle API errors with a focused message.
- Branch on `OutputFormat`.
- Print JSON or YAML through `io.PrintJSON` or `io.PrintYAML`.
- Render the table for the default path.
- Return data when that is useful for tests or for callers that need it.

For the organization list command, the method returns the fetched organizations after printing. That makes it easy to assert both the data and the output in tests.

## Output Pattern

If there is anything to output use the existing output conventions:

- Default: table output through `io.GetTableWriter()`.
- JSON: `io.PrintJSON(...)`.
- YAML: `io.PrintYAML(...)`.

When rendering tables, keep formatting close to the data shape:

- Build a header row that matches the columns the user sees.
- Convert optional fields before rendering.
- Keep special formatting in the command layer, not inside the API client.

## Registration Pattern

Add the new command to the correct parent command inside the command family folder.

For a folder-based command family, that usually means:

- Add the parent command file and all direct subcommand files in the same folder.
- Keep shared flag validation in the parent command.
- Pass only family-level shared dependencies from the parent to the children.
- Prefer a parent `PersistentPreRunE` to populate shared state before any child command runs.
- Set command-specific examples on the leaf command.
- If the command family needs to be attached to a higher-level command, do that from the family parent registration helper.

For example, the `start` family should be organized so the parent `start` command and the direct `pipeline` subcommand live together under `cli/cmd/start/`.

## Test Pattern

Match the existing command tests in `cli/cmd`.

### Test package layout

- Put tests in `cmd_test` when you want black-box coverage of the exported command API.
- Use `Ginkgo` and `Gomega` like the existing tests.
- Keep each command in its own `Describe` block.

### Test setup

Use the same fixture pattern as the organization list tests:

- Create `MockEnv` and `MockClient` in `BeforeEach`.
- Build the command wrapper with a populated `ListOptions` or the relevant options struct.
- Set `GlobalOptions.Env` so client creation can be exercised deterministically.
- Override `ClientFactory` when the command needs to bypass real client construction.
- Use `AfterEach` to assert expectations on the mocks.

### What to test

Cover the command in layers:

- `RunE` success path.
- `RunE` client creation failure.
- `RunE` API failure.
- The action method with successful data retrieval.
- JSON output path.
- YAML output path.
- Cobra registration, including `Use`, `Short`, and `RunE`.

### Stdout capture for JSON and YAML

When verifying printed output, follow the current approach:

- Replace `os.Stdout` with a pipe.
- Run the action method.
- Close the writer.
- Read the captured buffer.
- Unmarshal JSON with `encoding/json` or YAML with `go.yaml.in/yaml/v2`.
- Compare the decoded value with the expected slice.

## Recommended Edit Order

1. Create the command family folder under `cli/cmd`.
2. Add the parent command file and direct subcommand files to that folder.
3. Register the family to the top-level parent command.
4. Add or extend the client interface if the new command needs a new API call.
5. Add tests next to the command files in the same folder.
6. Run `go test ./cli/cmd`.
7. Run `make docs` if the new command changes generated CLI documentation.

## Practical Rules

- Keep `RunE` thin.
- Keep API logic out of registration helpers.
- Keep output formatting in the command layer.
- Prefer the smallest possible method surface that still allows isolated tests.
- Mirror the surrounding command’s naming style, but favor exported registration helpers when tests live in `cmd_test`.
- Keep new command families colocated so the folder name, command name, and test files tell the same story.