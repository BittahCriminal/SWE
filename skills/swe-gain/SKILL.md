---
name: swe-gain
description: >
  Explain the evidence behind SWE's minimal-engineering rules without inventing
  per-repository savings. Use for /swe-gain or "what does SWE save".
---

# SWE Gain

SWE does not claim universal percentage savings. When invoked, explain how to
measure SWE on the current workload: compare equivalent baseline and SWE runs,
then report changed lines, tokens, cost, elapsed time, test outcomes, and safety
checks. Report only measurements the user supplies or that can be reproduced
from available artifacts.

Never invent savings for the current repository—the unbuilt baseline does not
exist. Point users to `/swe-debt` for counted shortcuts and `/swe-audit` for
current deletion opportunities. One-shot display; change no mode or files.
