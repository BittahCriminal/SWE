---
name: swe-audit
description: >
  Quality audit of a whole repo: bugs, security holes, what breaks under real
  load, risky code without tests, slow paths, and what to delete, merge, or
  split. Ranked, with every finding explained in plain English. One-shot,
  read-only report. Use for "audit this codebase", "review the whole repo",
  "find bloat", "what can I delete", or /swe-audit.
---

# SWE Audit

Audit the whole repo like the senior developer who just inherited it and will
be paged when it breaks. Order of importance: correct, safe, holds under load,
tested, fast, lean. Lean still matters because every extra line must be read,
tested, and fixed later. Give the requested report in full.

## 1. Map first

- Audit what the user names: a folder, package, or whole repo. Nothing named:
  the whole repo.
- Read the README, deploy/build config, dependencies, entry points, and tests.
- Find expected load. Judge scale against it and state the assumption.
- Trace main flows end to end. Fully read paths handling user input, money,
  auth, data writes, background jobs, or state shared between processes.
- In a big repo, go deep where a mistake costs most. Say what you did not read.

## 2. Look for

1. **Bug:** wrong result, crash, missed edge case, callers disagreeing with a
   return value, or one rule applied differently in two places.
2. **Risk:** security holes or data loss, including swallowed errors, wrong
   write order, and missing transactions.
3. **Scale:** races, work repeated by every process, unbounded memory, a query
   per item, O(n²), or process-local state that must be shared.
4. **Missing test:** risky logic with no test that fails when it breaks.
5. **Speed:** big slowdowns are problems; small hot-loop wins are suggestions.
6. **Lean:** delete dead or speculative code; reuse duplicate helpers; prefer
   stdlib/native behavior; remove one-product abstractions; merge near-copies;
   split unrelated jobs, never merely by line count.

## 3. Check before reporting

- Every finding needs a concrete case. No case, no finding.
- Before calling code unused, grep the whole tree, including tests, fixtures,
  config, and string or dynamic references.
- A `shortcut:` or older `swe:` comment that names its limit is a decision, not
  a finding, unless expected load already crosses it.
- Propose the smallest fix, prefer deletion, and never add unneeded layers.
- No style taste, “consider”, or vague worries.

## 4. Output

Start with `What this repo does:` in two or three plain-English sentences and
name the assumed load. Then give at most 20 numbered findings, most important
first, grouped under **Must fix**, **Should fix**, and **Nice to have**. Skip
empty groups. Each finding has a bold title with `file:lines`, then four short
bullets: `What this is`, `Problem`, `Fix`, and `If we skip it`.

End with `Verdict:` naming what to fix first. Add
`Lean: -<N> lines, -<M> dependencies possible.` when lean findings exist and
`Not checked:` for unread or un-runnable parts.

Nothing found: `What this repo does:`, then `Healthy. Nothing to fix.` and one
line on what was checked. One-shot report; change no code.
