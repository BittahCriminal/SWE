# SWE

SWE is a Go-native smallest-correct-change runtime for AI coding agents,
delivered as one static binary with no runtime dependencies.

The important split is intentional:

- Go handles mode state, configuration, lifecycle hooks, host detection, and output protocols.
- Markdown holds the agent instructions and skills because that is what coding agents consume.

This keeps the lifecycle runtime native to Go without pretending that
instruction assets should become Go source.

## What it does

Before code is written, SWE walks a short ladder: skip speculative work, reuse
what exists, prefer the standard library, prefer native platform behavior,
reuse installed dependencies, and only then write the minimum implementation.
It does not trade away validation, data-loss protection, security,
accessibility, or a runnable check for non-trivial logic.

Modes:

| Mode | Behavior |
|------|----------|
| `lite` | Build the request and mention the lazier alternative. |
| `full` | Enforce the complete ladder. Default. |
| `ultra` | Prefer deletion and challenge speculative requirements. |
| `review` | Review only for over-engineering. Session-only. |
| `off` | Inject nothing. |

## Install

Go 1.22 or newer is required to build the binary:

```bash
go install github.com/bittahcriminal/SWE/cmd/swe@latest
```

Make sure the Go bin directory is on the non-interactive shell `PATH`; hooks
invoke `swe` directly.

### Codex

```bash
codex plugin marketplace add bittahcriminal/SWE
codex plugin add swe@swe
```

### Claude Code

```bash
/plugin marketplace add bittahcriminal/SWE
/plugin install swe@swe
```

### GitHub Copilot CLI

```bash
copilot plugin marketplace add bittahcriminal/SWE
copilot plugin install swe@swe
```

Copilot installs SWE from its GitHub-hosted marketplace and loads its skills
and native lifecycle hooks.

### Gemini CLI

```bash
gemini extensions install https://github.com/bittahcriminal/SWE --auto-update
```

Gemini installs SWE as an extension, including its skills, `/swe` and
`/swe-review` commands, and session-start context hook.

### OpenCode

```bash
swe install opencode
```

The Go installer adds a managed SWE block to the global OpenCode `AGENTS.md`
and installs SWE-owned skills and commands under `~/.config/opencode/`. It
preserves unrelated rules and is safe to run again when updating SWE. Set
`XDG_CONFIG_HOME` to use a different configuration root.

Start a new host session after installation. Qoder users can copy the template
under [`hooks/qoder-hooks.json`](hooks/qoder-hooks.json).

Instruction-only agents can copy [`AGENTS.md`](AGENTS.md) into a repository or
user-level agent configuration.

## CLI

```text
swe instructions [lite|full|ultra|review]  print the embedded ruleset
swe mode [MODE]                            read or set host-scoped session mode
swe default [MODE]                         read or persist the startup mode
swe status [--badge]                       inspect state or print [SWE:MODE]
swe install opencode [--config-dir PATH]   install global OpenCode integration
swe hook session-start                     lifecycle hook entrypoint
swe hook prompt-submit                     lifecycle hook entrypoint; reads JSON
swe hook subagent-start                    lifecycle hook entrypoint; reads JSON
```

Hook-based hosts accept `/swe lite|full|ultra|off`, `/swe default <mode>`, and
`/swe-review`. Gemini and OpenCode expose turn-scoped `/swe
[lite|full|ultra]` and `/swe-review` commands. Codex skill invocations use
`@swe` and `@swe-review`. The standalone phrases `stop swe` and `normal mode`
deactivate hook-based sessions.

## Configuration

Resolution order for the startup default:

1. `SWE_DEFAULT_MODE`
2. `defaultMode` in `~/.config/swe/config.json`
3. `full`

On Windows the config file lives at `%APPDATA%\swe\config.json`; setting
`XDG_CONFIG_HOME` overrides the base directory on every platform.

Session state is isolated by host: Codex uses `PLUGIN_DATA`, Copilot uses
`COPILOT_PLUGIN_DATA`, Gemini uses `~/.gemini`, Qoder uses `~/.qoder`, and
Claude uses `CLAUDE_CONFIG_DIR` or `~/.claude`.

Set `SWE_SUBAGENT_MATCHER` to a case-insensitive regular expression to limit
subagent injection by `agent_type`. Unset means every subagent. Invalid or
missing metadata fails open and injects the rules so the mode is never silently
lost.

## Scope and compatibility

The Go runtime covers the portable core:

- Claude Code, Codex, Copilot CLI, Gemini CLI, and Qoder hook output formats
- default and session mode persistence
- mode commands and standalone deactivation
- subagent context propagation and optional matching
- embedded main and review rules
- installable skills for review, audit, debt, gain, and help
- a cross-platform status badge via `swe status --badge`
- a Go-powered global installer for OpenCode rules, skills, and commands

OpenCode loads SWE through its native rules, skills, and command directories;
it does not expose SWE's lifecycle mode state. Pi and Hermes require
host-specific JavaScript or Python plugin APIs, so use
`swe instructions <mode>` as their host-neutral integration point.

## Development

```bash
gofmt -w .
go test ./...
go test -race ./...
go vet ./...
go build ./cmd/swe
```

The implementation uses only the Go standard library. CI tests Linux, macOS,
and Windows.

## License

SWE is an independent project released under the MIT License. See
[`NOTICE`](NOTICE) for project status and [`LICENSE`](LICENSE) for the license
terms and retained copyright notices.
