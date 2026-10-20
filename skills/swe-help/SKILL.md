---
name: swe-help
description: >
  Quick reference for SWE levels, skills, and commands. One-shot display.
  Use for /swe-help, "SWE help", or "how do I use SWE".
---

# SWE Help

Display this reference. One shot: do not change mode or persist anything.

## Levels

| Level | Trigger | What changes |
|-------|---------|--------------|
| **Lite** | `/swe lite` | Build what was asked and name the smaller option in one line. |
| **Full** | `/swe` | Make the smallest complete change, leave a check where logic needs one, and name skipped work and risk. Default. |
| **Ultra** | `/swe ultra` | Also question the request and push back before building. |

The level sticks until changed or the session ends.

## Skills

| Skill | What it does |
|-------|--------------|
| `/swe` | Lazy mode: the smallest complete change. |
| `/swe-review` | Quality review of a change: bugs, security, load, tests, speed, and bloat. |
| `/swe-audit` | The same quality review for the whole repo, ranked. |
| `/swe-debt` | Harvest `shortcut:` comments into a debt ledger. |
| `/swe-gain` | Explain how to measure SWE without inventing savings. |
| `/swe-help` | This card. |

In Codex, invoke namespaced skills such as `$swe:swe-review`.

## Deactivate and configure

Say "stop SWE" or "normal mode", or use `/swe off`. Resume with `/swe`.
Default mode is `full`; override it with `SWE_DEFAULT_MODE` or
`~/.config/swe/config.json` (`%APPDATA%\swe\config.json` on Windows):

```json
{ "defaultMode": "lite" }
```

Resolution order: environment, config file, then `full`.
