---
name: swe-review
description: >
  Review a diff only for over-engineering. Find code to delete, standard-library
  or native replacements, speculative abstractions, and shorter equivalent logic.
  Use for "review for over-engineering", "what can we delete", or /swe-review.
---

# SWE Review

Review the target for unnecessary complexity. One line per finding; the best
outcome is a shorter diff.

## Format

`<file>:L<line>: <tag> <what to cut>. <replacement>.`

Tags:

- `delete:` dead code, unused flexibility, or a speculative feature. Replacement: nothing.
- `stdlib:` hand-written behavior already supplied by the language. Name it.
- `native:` a dependency or custom code replaced by a platform feature. Name it.
- `yagni:` an interface with one implementation, unused option, or layer with one caller.
- `shrink:` equivalent behavior in fewer, clearer lines. Show the form.

Rank findings by impact. End with `net: -<N> lines possible.` If there is
nothing material to cut, say `Lean already. Ship.` and stop.

Correctness, security, and performance bugs belong in a normal review pass.
Do not apply fixes unless asked. Never flag the single runnable check required
for non-trivial behavior as bloat.
