# SWE

SWE is a Go-native lazy senior developer mode for AI coding agents. It makes
the smallest complete change: every caller, test, fixture, config entry, and
export the task really touches, with no speculative layers around it.

The split is intentional:

- Go handles mode state, configuration, lifecycle hooks, host protocols, and a
  bounded codebase map.
- Markdown holds instructions and skills because that is what agents consume.

One static binary runs the portable lifecycle; there are no runtime package
dependencies.

## What it does

Before writing, SWE traces the real path a change touches and checks what it
could break. It then takes the first complete option that works: skip an
unneeded feature, reuse code already present, use the standard library or
platform, reuse an installed dependency, use one clear line, or finally write
the minimum implementation.

SWE never cuts trust-boundary validation, data-loss protection, security,
accessibility, real-hardware calibration, requested behavior, or the one small
test non-trivial logic needs. Replies end with skipped work and known risk.

| Mode | Behavior |
|------|----------|
| `lite` | Build what was asked and name the smaller option. |
| `full` | Make the smallest complete change. Default. |
| `ultra` | Also challenge unjustified parts of the request before building. |
| `off` | Inject nothing. |

The bundled one-shot skills are `swe-review`, `swe-audit`, `swe-debt`,
`swe-gain`, and `swe-help`. Review and audit now cover correctness, security,
real load, missing tests, speed, and lean code—not only over-engineering.

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

Review and trust the lifecycle hooks, then start a new task.

### Claude Code

```text
/plugin marketplace add bittahcriminal/SWE
/plugin install swe@swe
```

### GitHub Copilot CLI

```bash
copilot plugin marketplace add bittahcriminal/SWE
copilot plugin install swe@swe
```

Copilot loads the six skills and native session hooks.

### Gemini CLI

```bash
gemini extensions install https://github.com/bittahcriminal/SWE --auto-update
```

Gemini loads `AGENTS.md` as always-on context and registers all six `/swe`
commands and skills. Commands are turn-scoped; Gemini does not load the
Claude/Codex lifecycle configuration.

### OpenCode

```bash
swe install opencode
```

The Go installer merges a managed SWE block into the global OpenCode
`AGENTS.md` and installs all six skills and commands under
`~/.config/opencode/`. Existing rules are preserved, and rerunning the command
updates only SWE-owned content. Set `XDG_CONFIG_HOME` to choose another root.

Start a new host session after installation. Qoder users can copy the template
under [`hooks/qoder-hooks.json`](hooks/qoder-hooks.json). Instruction-only
agents can use [`AGENTS.md`](AGENTS.md) directly.

## CLI

```text
swe instructions [lite|full|ultra|review]  print an embedded ruleset
swe mode [MODE]                            read or set lite/full/ultra/off
swe default [MODE]                         read or persist the startup mode
swe status [--badge]                       inspect state or print [SWE:MODE]
swe install opencode [--config-dir PATH]   install global OpenCode integration
swe hook session-start                     lifecycle hook entrypoint
swe hook prompt-submit                     lifecycle hook entrypoint; reads JSON
swe hook subagent-start                    lifecycle hook entrypoint; reads JSON
```

Hook-based hosts accept `/swe lite|full|ultra|off`, `/swe default <mode>`, and
the standalone phrases `stop swe` and `normal mode`. Bare `/swe` reports the
live level or reactivates SWE at the configured default (`full` when the
default is `off`). Review is one-shot and never becomes a session mode.

Gemini and OpenCode expose turn-scoped `/swe`, `/swe-review`, `/swe-audit`,
`/swe-debt`, `/swe-gain`, and `/swe-help`. Codex skill invocations use the
plugin namespace, for example `$swe:swe-review`.

## Configuration

Startup default resolution:

1. `SWE_DEFAULT_MODE`
2. `defaultMode` in `~/.config/swe/config.json`
3. `full`

On Windows the config file is `%APPDATA%\swe\config.json`.
`XDG_CONFIG_HOME` overrides the base directory on every platform.

Live mode is isolated by host and, when the host supplies a project directory,
by project. A shared flag is also maintained for status lines. Context
compaction preserves the live level instead of resetting to the default.
Qoder honors `QODER_CONFIG_DIR` and remembers an explicit `off` across prompts.

Set `SWE_SUBAGENT_MATCHER` to a case-insensitive regular expression to limit
subagent injection by `agent_type`. Missing metadata and invalid expressions
fail open so the rules are not silently lost.

Session start also adds a bounded map of reusable top-level symbols. It reads
tracked source files, skips tests, generated/build/vendor paths, unsafe names,
and symlinks, and emits identifiers only. Set `SWE_MAP=0` to disable it or
`SWE_MAP_CHARS` to change the default 2000-character budget.

## Scope

The Go runtime provides:

- Claude Code, Codex, Copilot CLI, Gemini CLI, and Qoder output formats
- default, host, and project mode persistence
- mode switching, standalone deactivation, and compaction preservation
- subagent propagation and optional matching
- safe, bounded codebase mapping
- six installable skills and host-native command files
- an OpenCode rules/skills/commands installer
- a cross-platform status badge through `swe status --badge`

OpenCode and Gemini use native instruction, skill, and command loading rather
than SWE lifecycle state. Hosts that require JavaScript or Python plugin APIs
can use `swe instructions <mode>` as the Go-neutral integration point.

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
[`NOTICE`](NOTICE) and [`LICENSE`](LICENSE).
