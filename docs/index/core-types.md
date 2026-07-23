---
id: okf-002
feature: core-types
branch: feature/core-types
status: done
files:
  - pii/types.go (Result, Pattern, Explanation, MaxScore)
  - nlp/nlp.go (minimal Artifacts — the engine interface arrives with nlp-onnx)
  - recognizer/recognizer.go (Recognizer interface, ValidateFunc)
  - recognizer/pattern_recognizer.go (NewPattern, WithValidate, rune offsets)
  - registry/registry.go (Add/Remove/Get(language, entities...)/SupportedEntities, RWMutex)
  - examples/basic/main.go (real E2E)
tests:
  - recognizer/pattern_recognizer_test.go (100% coverage)
  - registry/registry_test.go (100% coverage)
decisions:
  - "2026-07-16: ValidateFunc returns *bool — nil neutral, true→MaxScore, false→rejected (modeled on the fork's validate/invalidate_result)"
  - "2026-07-16: bytes→runes conversion via utf8.RuneCountInString at the match point (no precomputed table until measured)"
  - "2026-07-16: Registry protected by RWMutex — mutable setup, concurrent Analyze"
---

**What**: the foundation — shared types, Recognizer interface,
PatternRecognizer (regex + checksum validation), Registry filtering by
language/entity. Rune offsets verified on accented text (email-rune-offsets
oracle case).

**Pitfalls**:
- `go test -race` requires cgo/gcc, absent on Windows by default → -race
  delegated to Linux CI.
