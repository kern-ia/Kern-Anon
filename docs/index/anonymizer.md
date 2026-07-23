---
id: okf-004
feature: anonymizer
branch: feature/anonymizer
status: done
files:
  - anonymizer/operators.go (Replace, Redact, Keep, Custom, Mask, Hash, Encrypt/Decrypt AES-GCM)
  - anonymizer/engine.go (Engine.Anonymize/Deanonymize, resolveConflicts, Item/Result)
tests:
  - anonymizer/anonymizer_test.go (12 tests, 90.9% — overlaps, runes, round-trip)
decisions:
  - "2026-07-16: operators = closures behind a 2-method interface (no factory — idiomatic Go, replaces OperatorsFactory)"
  - "2026-07-16: AES-GCM base64(nonce||ct) encryption instead of the fork's AES-CBC (authenticated, Go-only round-trip)"
  - "2026-07-16: overlaps = score desc, length desc, position asc, greedy selection with no intersection"
  - "2026-07-16: Item exposes offsets (runes) in the OUTPUT text — needed for Deanonymize"
---

**What**: the anonymization engine — operator chosen per entity, then
DEFAULT, then Replace("<ENTITY_TYPE>"); overlap resolution before
substitution; Deanonymize replays the items (typically Decrypt) for the
round-trip.

**Pitfalls**: nothing new — the only TDD red was an arithmetic error in a
test expectation (16−12=4 digits remaining, not 6).
