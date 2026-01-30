# AGENTS Guidelines

## Branching & Commits
- Branch from `main` for new work.
- Follow the Gitflow workflow for branching and releases.
- Use imperative, present-tense commit messages (e.g. `Add feature X`).
- Use Conventional Commits for commit messages.

## Preferred Workflow
- Use the locally available Go version (as specified in `go.mod`).
- Run `go tool golangci-lint run ./...` before pushing.
- Ensure `go test ./...` passes.

## Runtime Requirements
- Validate `Invocation.Input` (or `input.json` when `Input` is nil) against `InputSchema` before running the agent.
- `RunDir` must exist; the runner should not create it implicitly.
- Stdout/stderr streaming is opt-in via options; omit them to disable streaming output.
- On successful runs (exit code 0), the agent must write `output.json` in `RunDir` that conforms to `OutputSchema`.
- On successful runs, return an error if the output file is missing or does not match `OutputSchema`.

## CLI Requirements
- Provide a CLI `exec` command that runs arbitrary agent commands with normalized JSON I/O.
- Provide predefined agent commands (e.g., `codex`) that map to the corresponding agent command plus any extra args.
- The CLI must emit the agent's output JSON to stdout after a successful run while preserving the agent's exit code.
- Flags must support: `--input-schema`, `--output-schema`, `--prompt`, `--input`, `--extra-args`, and `--work-dir`.
- Flags must also support: `--input-schema-file` and `--output-schema-file`.
- Flags must also support: `--debug` to forward agent stdout/stderr to stderr.
- `--input-schema` and `--output-schema` default to `{"type":"string"}` when omitted.
- `--input` sets `Invocation.Input` (string value) and drives `input.json` creation.
- `exec` must support: `--tty` to run the agent in a pseudo-terminal.
- `codex`, `opencode`, `gemini`, and `claude` must support: `--model` (optional; can also be passed via `--extra-args`).
- The `codex` command must enforce `exec` mode.

## Agent Development Kit (ADK)
- Provide a high-level `ExecAgent` that wraps the runner and implements the `agent.Agent` interface.
- Constructors (e.g., `NewExecAgent`) must return concrete types (e.g., `*ExecAgent`) while accepting interfaces where appropriate.
- Functional options MUST be used for configuration, powered by `options-gen`.
- Generated options MUST include validation and be called in the constructor.
- `ExecAgent` must support schema overrides, system prompts, extra arguments, and custom `RunDir`.
- Agents built using the ADK SHOULD use the standard `launcher` package (`google.golang.org/adk/cmd/launcher`) to provide a CLI interface that is fully compatible with `ainvoke`.

## Code Style (Google Go Style)
- Adhere to the [Google Go Style Guide](https://google.github.io/styleguide/go/).
- **Naming**: Avoid repetition (e.g., `user.Type` not `user.User`). Use noun-like names for getters.
- **Errors**: Wrap with `%w` only if programmatic inspection is needed. Avoid redundant logging.
- **Testing**: Use table-driven tests. Ensure failures are actionable and descriptive.
- **Documentation**: Focus on "why" for non-obvious logic. Document concurrency safety.
