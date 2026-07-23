---
id: okf-006
feature: analyzer
branch: feature/analyzer
status: done
files:
  - analyzer/analyzer.go (Engine, options construction + call options)
  - nlp/nlp.go (Engine interface + NoOp — ready for nlp-onnx)
  - examples/basic/main.go (E2E: public API from the plan, full pipeline)
tests:
  - analyzer/analyzer_test.go (8 tests, 88.4% — pipeline, filters, dedup, NLP wiring)
decisions:
  - "2026-07-16: functional call options (Language/Entities/MinScore) — plan API §4 implemented as-is"
  - "2026-07-16: dedup = span contained in a span of the SAME type with score ≥ ; cross-entity overlaps remain (as in the fork) and are resolved by the anonymizer"
  - "2026-07-16: context enhancer enabled by default, WithEnhancer(nil) to disable it"
  - "2026-07-16: nlpEngine.Load() called in New — load failure = construction error, no call made"
---

**What**: the engine orchestrating the pipeline — NLP (optional) → registry
recognizers → context boost → same-entity dedup → threshold → sort by
position. This is the library's entry API.

**Pitfalls**: none new — NLP wiring is verified by a fake recognizer that
captures the artifacts.
