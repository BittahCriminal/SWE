---
name: swe-audit
description: >
  Audit an entire repository for over-engineering and rank what can be deleted,
  simplified, or replaced with standard-library or native behavior. Read-only.
---

# SWE Audit

Run `swe-review` across the whole tree. Look for redundant dependencies,
single-implementation interfaces, one-product factories, delegating wrappers,
dead flags, unused configuration, and hand-written standard-library behavior.

Report one line per finding, largest cut first:

`<tag> <what to cut>. <replacement>. [path]`

Use `delete`, `stdlib`, `native`, `yagni`, and `shrink` tags. End with
`net: -<N> lines, -<M> dependencies possible.` If nothing is worth cutting,
say `Lean already. Ship.` Do not change files.
