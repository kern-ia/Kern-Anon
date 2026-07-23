---
id: okf-005
feature: context-enhancer
branch: feature/context-enhancer
status: done
files:
  - contextaware/enhancer.go (Enhancer, fork defaults 0.35/0.4/5-before/0-after)
  - recognizer/pattern_recognizer.go (WithContextWords, ContextWords())
  - recognizers/generic/*.go (fork CONTEXT words added to the 7 recognizers)
tests:
  - contextaware/enhancer_test.go (7 tests: boost, window, min, cap)
decisions:
  - "2026-07-16: \"substring\" matching on the joined lowercase window (fork's mode) — no lemmatization until NLP is wired in"
  - "2026-07-16: originating recognizer retrieved via Explanation.Recognizer; ContextAware interface via type assertion"
  - "2026-07-16: context words injected via a Python script (7 files at once, count assertions)"
---

**What**: context boost — if a recognizer's context word appears within the
5 words preceding the entity, score +0.35 (min 0.4, capped at 1.0),
ContextBoost recorded in the Explanation.

**Pitfalls**:
- Substring matching will boost "ip" in "description": faithful to the fork,
  to revisit once NLP lemmatization is wired in.
