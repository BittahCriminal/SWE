---
name: swe-debt
description: >
  Collect every deliberate `swe:` shortcut comment into a debt ledger with its
  ceiling and upgrade trigger. Use for /swe-debt or "list SWE shortcuts".
---

# SWE Debt

Search comments for `swe:` while skipping version-control, dependency, and
build-output directories. Each marker should use
`swe: <known ceiling>, <upgrade trigger>`.

Group results by file:

`<file>:<line>, <simplification>. ceiling: <limit>. upgrade: <trigger>.`

Tag markers without a concrete upgrade condition as `no-trigger`. End with
`<N> markers, <M> with no trigger.` If none exist, say
`No swe: debt. Clean ledger.` Read and report only unless asked to persist it.
