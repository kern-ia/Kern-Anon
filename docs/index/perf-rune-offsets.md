---
id: okf-011
feature: perf-rune-offsets
branch: feature/perf-rune-offsets
status: done
files:
  - recognizer/pattern_recognizer.go (collect bytes then runeOffsets in a single pass)
tests:
  - recognizer/pattern_recognizer_bench_test.go (60 KB accented text, 800 matches)
  - existing suite unchanged (identical offsets)
decisions:
  - "2026-07-17: single-pass conversion via `for byteIdx := range text` (iterates rune starts) — regex boundaries are always rune boundaries (RE2)"
  - "2026-07-17: benchmarked BEFORE optimizing to establish a baseline (25.3 ms/op)"
---

**What**: removal of the quadratic behavior in bytes→runes conversion —
per-match `utf8.RuneCountInString` replaced with a single pass over the text
for all matches.

**Measurement**: 25.3 ms/op → 2.39 ms/op (×10.6), 2.6 → 27.9 MB/s, same text
(60 KB, 800 matches). No behavior change (suite green).
