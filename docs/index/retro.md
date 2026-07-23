# Ongoing retro — presidigo-go

## 2026-07-17 — v0.2.0 retro (performance)
**Key lesson**: the baseline benchmark invalidated the intuition — the
bottleneck wasn't recognizer sequentiality but the context window
(O(results × text)): 550 ms and 534 MB allocated on 30 KB. Always measure
BEFORE optimizing. Gains delivered: recognizer ×10.6 (single-pass rune
offsets), pipeline ×41 (bounded window + fan-out), windowed NER 256/64 in
parallel (long texts finally covered, better labels than at 512). Go/Python
harness: still 100%.

## 2026-07-17 — v0.1.0 final retro
**What worked**: the oracle corpus (Python tests → shared jsonl) carried
all of TDD; AST/sed extraction of the Python patterns avoided typos; the
Go-vs-Python harness immediately caught the one real porting gap (SSN
invalidation) → final 100% agreement; the 3 RE2 pitfalls anticipated in the
plan are exactly the ones encountered.
**What cost time**: onnxruntime ↔ binding version mismatches (2 attempts),
a cased model lowercased (zero entities, no error), cp1252 encoding of the
.ps1 scripts.
**Next up**: EU locales (AST porting already tooled), lemmatization in
contextaware once NLP provides lemmas, PHONE_NUMBER via
nyaruka/phonenumbers, publish the module on GitHub.

Record here what worked or didn't, AS IT HAPPENS. Date every entry.

## 2026-07-16 — bootstrap
- Go missing from the machine: installed via `winget install GoLang.Go`
  (go1.26.5). Remember to refresh the PowerShell session PATH after winget.

## 2026-07-16 — core-types
- `go test -race` on Windows requires cgo (gcc): not installed → run local
  tests without -race, Ubuntu CI handles it.
- Rune offsets: `FindAllStringIndex` returns bytes; converted at the match
  point with `utf8.RuneCountInString` — tested with "Prénom/José"
  (multi-byte).

## 2026-07-16 — generic-recognizers
- As anticipated in the plan: 3 non-RE2 Python regexes (card lookahead, IP
  lookbehind, MAC backreference). Fixes: Go validation, netip, split
  patterns.
- Large regex (URL, ~6 KB of TLDs): extracted by script from the Python
  source rather than retyped — zero typo risk, regenerable.
- The IDE's gopls points at the presidigo (Python) workspace: phantom
  errors in presidigo-go → open presidigo-go as a separate project or add
  a go.work.
