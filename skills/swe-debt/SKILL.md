---
name: swe-debt
description: >
  List every `shortcut:` comment and older `swe:` marker as a debt ledger.
  One-shot, read-only report. Use for "SWE debt", "what did SWE defer",
  "list the shortcuts", or /swe-debt.
---

# SWE Debt

Every deliberate shortcut should name its ceiling and upgrade path. This skill
collects those decisions so a deferral cannot quietly become permanent.

## Scan

Search tracked and unignored files for comment markers:

`git grep --no-index --exclude-standard -nE '(#|//|/[*]) ?(shortcut|swe):'`

Add other comment prefixes when the stack needs them. If the user names a
custom marker (`/swe-debt TODO`), search for that instead. Skip prose and notes
about keyboard shortcuts that are not deferrals.

## Output

One row per marker, grouped by file:

`<file>:<line>, <what was simplified>. ceiling: <limit>. upgrade: <trigger>.`

The convention is `shortcut: <ceiling>, <upgrade path>`. Flag a marker without
a concrete upgrade path as `no-trigger`. End with
`<N> markers, <M> with no trigger.` Nothing found:
`No shortcut debt. Clean ledger.`

Read and report only. Persist a ledger only when asked.
