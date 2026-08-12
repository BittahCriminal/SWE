---
name: swe-help
description: >
  Quick reference for SWE modes, skills, configuration, and deactivation.
  Use for /swe-help, "SWE help", or "what SWE commands exist".
---

# SWE Help

Show this compact reference:

| Command | Purpose |
|---------|---------|
| `/swe lite` | Build the request and mention the lazier alternative. |
| `/swe` or `/swe full` | Enforce the full minimal-engineering ladder. |
| `/swe ultra` | Challenge speculative work and prefer deletion. |
| `/swe off` | Disable the mode for the session. |
| `/swe default <mode>` | Persist the startup mode. |
| `/swe-review` | Review the current diff for over-engineering. |
| `/swe-audit` | Audit the repository for deletion opportunities. |
| `/swe-debt` | Collect deliberate `swe:` shortcuts. |
| `/swe-gain` | Show the attributed upstream benchmark. |

Say "stop swe" or "normal mode" to deactivate. Configure the default with
`SWE_DEFAULT_MODE` or `~/.config/swe/config.json`. Codex invokes skills with
`@swe`, `@swe-review`, and the corresponding names.
