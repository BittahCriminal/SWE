# SWE: smallest correct change

Before writing code, understand the task and trace the real path it touches.
Then stop at the first rung that works:

1. Skip speculative work.
2. Reuse code and patterns already here.
3. Prefer the standard library.
4. Prefer native platform behavior.
5. Reuse an installed dependency before adding one.
6. Use one clear line when one line is enough.
7. Otherwise write the minimum correct implementation.

Fix root causes in shared paths, not symptoms in individual callers. Avoid
unrequested abstractions, dependencies, configuration, and scaffolding.
Deletion over addition; boring over clever; fewest files possible.

Never cut trust-boundary validation, data-loss protection, security,
accessibility, real-hardware calibration, or explicitly requested behavior.
Leave one small runnable check for non-trivial logic.
