---
name: swe-review
description: >
  Quality review of a change: is the logic right, is it safe, does it hold
  under real load, is risky code tested, is it fast enough, and is every line
  needed. Reads connected code, not only the diff. Each finding is explained
  in plain English. Use for "review this", "code review", "review my PR",
  "is this over-engineered", or /swe-review.
---

# SWE Review

Review a change like the senior developer who will be paged when it breaks.
Order of importance: correct, safe, holds under load, tested, fast, lean. Lean
still matters: every extra line must be read, tested, and fixed later. This is
a report the user asked for, so give it in full.

## 1. Understand first

- Review what the user names: uncommitted or staged changes, a branch, a PR
  link, or files. Nothing named: uncommitted changes, or the last commit if
  there are none.
- Read the diff, then the code it touches: callers of every changed function,
  functions it calls, tests, and README.
- Trace the real flow: where data comes in, what is stored, what goes out.
- A change can break untouched code. When a signature, return value, or
  behavior changes, grep every caller.
- Find expected load in the repo: one person running a script, or many users
  and processes. Judge scale against that and state the assumption.

## 2. Look for

1. **Bug:** wrong result, crash, missed edge case, caller broken by the change,
   or a fix applied in one caller while shared code stays broken.
2. **Risk:** injection, weak randomness, secrets, unchecked user input, errors
   swallowed, writes in the wrong order, or a missing transaction.
3. **Scale:** check-then-write races, work repeated by every process, unbounded
   memory, a query per item, O(n²) on large input, or process-local state that
   must be shared.
4. **Missing test:** risky new logic—a branch, parser, money, security, data
   writes, or a bug fix—with no test that fails when it breaks.
5. **Speed:** big slowdowns are problems. Small hot-loop wins are suggestions.
6. **Lean:** code that should not exist or should be smaller.
   - delete: dead code, unused options, speculative features
   - reuse: the repo already has this helper; name its path
   - stdlib/native: the platform already does it, or a dependency replaces a few lines
   - yagni: one-implementation abstraction or config nobody sets
   - merge: near-copies that must change together
   - split: one function doing unrelated jobs; split by job, never line count

## 3. Check before reporting

- Every finding needs a concrete case: this input or situation causes this
  result. No case, no finding.
- Re-read the lines and confirm the caller exists, the value can be empty, or
  the code is truly unused.
- A `shortcut:` or older `swe:` comment that names its limit is a decision, not
  a finding, unless expected load already crosses it.
- Propose the smallest fix. Prefer deletion. Never add unneeded layers.
- No style taste, “consider”, or vague worries.

## 4. Output

Use short sentences and plain English. Explain technical terms the first time.

Start with `What this change does:` in two or three sentences and name the load
assumption. Then number findings across these groups, skipping empty groups:

- **Must fix:** bug, security, data loss, expected-load failure.
- **Should fix:** risky code without a test, real slowness, duplication, mixed
  jobs, or code that should not exist.
- **Nice to have:** small speed-ups and shorter forms.

Each finding has a bold title with `file:lines`, then four short bullets:
`What this is`, `Problem`, `Fix`, and `If we skip it`.

End with `Verdict: Ship.` or `Verdict: fix 1 and 3 first.` Add
`Lean: -<N> lines possible.` when lean findings exist and `Not checked:` when
something important could not be checked.

Nothing found: `What this change does:`, then `Looks good. Ship.` and one line
on what was checked. List findings; change no code.
