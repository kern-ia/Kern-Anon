---
id: okf-009
feature: nlp-onnx
branch: feature/nlp-onnx
status: done
files:
  - nlp/bertner/tokenizer.go (pure-Go WordPiece, rune offsets + BIO Aggregate)
  - nlp/onnx/engine.go (nlp.Engine via onnxruntime, onnx build tag)
  - recognizers/ner/ner.go (artifacts → PII: PER→PERSON, LOC→LOCATION, ORG→ORGANIZATION, MISC→NRP)
  - nlp/nlp.go (NerEntity, Artifacts.NerEntities)
  - scripts/download-model.{ps1,sh} (quantized Xenova/bert-base-NER + onnxruntime 1.26.0)
  - examples/ner/main.go (E2E with onnx tag)
tests:
  - nlp/bertner/tokenizer_test.go (90.9% — embedded vocab, runes, BIO)
  - recognizers/ner/ner_test.go (95% — mapping, nil artifacts, unknown label)
decisions:
  - "2026-07-17: model = int8-quantized Xenova/bert-base-NER (~110 MB) — bert-base quality, contained size"
  - "2026-07-17: onnxruntime loaded dynamically (ONNXRUNTIME_LIB) — version 1.26.x required by yalue/onnxruntime_go v1.31 (API 26)"
  - "2026-07-17: CASED tokenizer by default (do_lower_case read from tokenizer_config.json) — lowercasing broke NER"
  - "2026-07-17: local gcc = winlibs (winget BrechtSanders.WinLibs.POSIX.UCRT) → cgo and -race available on Windows"
---

**What**: NER — pure-Go WordPiece tokenizer and BIO aggregation (testable
everywhere), BERT-NER ONNX inference behind the `onnx` tag, ner recognizer
mapping entities to PERSON/LOCATION/ORGANIZATION/NRP. Local E2E verified:
"John Smith"/"Microsoft"/"Seattle" detected at 1.00.

**Pitfalls**:
- onnxruntime ↔ yalue version: the API must match (v1.31 → ORT 1.26.x),
  otherwise "API version [26] is not available".
- Cased model: lowercasing the tokenizer input makes all entities vanish.
- Accented .ps1 without BOM → PowerShell 5.1 (ANSI) parse error: keep
  scripts in ASCII.
