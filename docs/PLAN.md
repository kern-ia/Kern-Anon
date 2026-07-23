# Redesign plan — presidigo-go
> Written 2026-07-16 — greenfield-tdd-okf workflow, Phase 1 (to be validated before any line of code)

## 1. Context & goal

Port the core of Presidio (PII detection + anonymization in **text**) to Go,
as an **importable library** (`go get github.com/YoLaub/PresidioGo`), meant to
become an in-process module of an existing Go project.

Validated motivation: eliminate the Python sidecar (~500 MB-1 GB RAM with
spaCy) and the HTTP round-trip; ~80% of the value (pattern/checksum
recognizers) is portable pure logic; regex performance is superior in Go.

The reference functional scope is `presidigo/archi-output/INDEX.md`.

## 2. Validated decisions (2026-07-16)

| Decision | Choice |
|---|---|
| v1 scope | Text only: analyzer + anonymizer (no image/DICOM, no structured data) |
| NER | Embedded ONNX (bert-base-NER converted to ONNX, via onnxruntime/cgo) behind a pluggable interface |
| Location | New repo: `c:\Users\laube\WebstormProjects\presidigo-go` |
| Go module | `github.com/YoLaub/PresidioGo` |
| API | Idiomatic Go (NO Presidio REST compatibility) |
| v1 locales | Generic → USA (port) → France (new) → EU in waves |

## 3. Open source references studied

| Project | Finding |
|---|---|
| [aliengiraffe/deidentify](https://github.com/aliengiraffe/deidentify) | Go, deterministic HMAC anonymization — good idea to reuse for the hash operator, but no extensible architecture |
| [aavaz-ai/pii-scrubber](https://github.com/aavaz-ai/pii-scrubber) | Go, predefined regex scrubbers — no registry, no scoring, no NER |
| Presidio Python (this fork) | The reference architecture: registry + recognizers + context enhancement + operators |

Conclusion: no existing Go library provides extensible registry + scoring +
NER + operators. The rewrite is justified; deidentify is used as inspiration
for deterministic hashing.

## 4. Architecture

```
presidigo-go/
├── pii/              # shared types (no internal dependencies)
├── recognizer/       # Recognizer interface + generic PatternRecognizer
├── registry/         # Registry: add/remove/filter by language & entity
├── recognizers/
│   ├── generic/      # creditcard(Luhn), email, iban(mod97), ip, url, crypto, phone…
│   ├── us/           # ssn, itin, passport, driver_license (ported from the fork)
│   ├── fr/           # nir(social security, key 97), siren/siret(Luhn), cni, license (to create)
│   └── ...           # de/, it/, es/, pl/… in waves
├── nlp/              # NlpEngine + NlpArtifacts interface; noop included
│   └── onnx/         # BERT-NER via onnxruntime (cgo) — `onnx` build tag
├── contextaware/      # ContextEnhancer: score boost from context words
├── analyzer/         # Engine: orchestrates nlp → registry → contextaware → thresholds
├── anonymizer/        # Operator interface + replace/mask/hash/redact/keep/encrypt
│                     # Engine.Anonymize + overlap resolution
└── internal/testdata # oracle corpus extracted from the fork's Python tests
```

### Target public API

```go
// Analysis
eng, _ := analyzer.New(
    analyzer.WithRegistry(registry.Default("en", "fr")),
    analyzer.WithNlpEngine(onnx.New("models/bert-ner.onnx")), // or nlp.NoOp{}
)
results, err := eng.Analyze(ctx, text,
    analyzer.Entities("EMAIL_ADDRESS", "FR_NIR"),
    analyzer.MinScore(0.4),
)

// Anonymization
anon := anonymizer.New()
out, err := anon.Anonymize(text, results, map[string]anonymizer.Operator{
    "DEFAULT":       anonymizer.Replace("<PII>"),
    "EMAIL_ADDRESS": anonymizer.Mask('*', 4),
})
```

### Mapping to the Python fork

| Python (INDEX.md) | Go |
|---|---|
| AnalyzerEngine.analyze() | analyzer.Engine.Analyze() |
| RecognizerRegistry | registry.Registry |
| EntityRecognizer / PatternRecognizer | recognizer.Recognizer (interface) / recognizer.Pattern |
| NlpEngine (spaCy/Stanza/…) | nlp.NlpEngine (onnx / noop) |
| LemmaContextAwareEnhancer | contextaware.Enhancer |
| Operators + OperatorsFactory | anonymizer.Operator (interface, no factory — idiomatic) |
| ConflictResolutionStrategy | anonymizer: sort + merge of overlapping spans |
| REGEX_TIMEOUT_SECONDS (anti-ReDoS) | unnecessary: Go's regexp package (RE2) runs in guaranteed linear time |

## 5. Data model

```go
package pii

type Result struct {
    EntityType  string   // "EMAIL_ADDRESS", "FR_NIR", …
    Start, End  int      // byte offsets AND an exposed rune-based version
    Score       float64  // 0.0–1.0
    Explanation *Explanation // recognizer, pattern, applied context boost
}

type Pattern struct {
    Name  string
    Regex *regexp.Regexp
    Score float64
}
```

Hard points identified from the start:
- **Offsets**: Python indexes in runes, Go in bytes → decision: rune-based API
  (compatible with the oracle corpus), single internal conversion.
- **Post-regex validation**: Luhn, IBAN mod-97, NIR key = `Validate(match) *bool`
  functions (nil = neutral, true = score→max, false = rejected), as in the fork.

## 6. Steps (feature branches, strict TDD)

Each step: tests written BEFORE the code → green suite → E2E → OKF note
`docs/index/<feature>.md` → `--no-ff` merge into dev.

| # | Branch | Content | Depends on |
|---|---|---|---|
| 1 | feature/bootstrap | go.mod, golangci-lint, GitHub Actions CI, CLAUDE.md, docs/index/ + retro.md, oracle corpus extraction from the Python tests | — |
| 2 | feature/core-types | pii.Result, Pattern, Recognizer interface, PatternRecognizer, Registry | 1 |
| 3 | feature/generic-recognizers | the ~10 generic recognizers with validators (Luhn, mod-97…) | 2 |
| 4 | feature/anonymizer | Operators + Engine + overlaps | 2 (parallel with 3) |
| 5 | feature/context-enhancer | boost from context words + window | 3 |
| 6 | feature/nlp-onnx | NlpEngine interface, ONNX engine (cgo, build tag), noop fallback | 2 (parallel with 3-5) |
| 7 | feature/recognizers-us | port SSN, ITIN, passport, driver license | 3 |
| 8 | feature/recognizers-fr | create NIR, SIREN/SIRET, national ID, driver license, plate | 3 |
| 9 | feature/recognizers-eu | DE, IT, ES, PL… in waves based on actual need | 3 |
| 10 | release v0.1.0 | full oracle E2E, final retro, dev→main merge, tag | all |

Parallelizable: (3, 4, 6) after the base 2; (7, 8, 9) after 3.

## 7. E2E verification

1. **Oracle corpus**: synthetic sentences + cases from the fork's Python tests
   (`internal/testdata/oracle.jsonl`: text → expected entities with offsets).
2. **Comparison harness**: script that runs the Python presidio-analyzer
   (the fork's docker compose, port 5002) and compares its `/analyze` outputs
   against the Go outputs on the same corpus. v0.1.0 criterion: ≥ 95%
   agreement on pattern entities; NER compared for reference only (different
   engines).
3. **Real E2E**: a small example Go program (`examples/`) imported the way
   your project would — compiles, detects, anonymizes, encrypt/decrypt
   round-trip.

## 8. Risks & mitigations

| Risk | Mitigation |
|---|---|
| cgo (onnxruntime) complicates the build | `onnx` build tag: without the tag, 100% pure Go library with nlp.NoOp; NER becomes opt-in |
| Rune/byte offset divergence | oracle corpus with accents/emoji from step 2 onward |
| Go's RE2 doesn't support lookahead/backreferences present in some Python regexes | rewrite those regexes (validation in Go code) — tracked during the port, noted in retro.md |
| Scope creep (66 locales) | waves on demand, each wave = one OKF note |

## 9. Out of scope for v1 (explicit)

Image/DICOM, structured data (CSV/JSON), HTTP server, batch engines, Azure
AHDS operator, multi NLP engines. Extensible later without breaking the API
(interfaces).
