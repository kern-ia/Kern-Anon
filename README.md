# PresidioGo

Go library for **detecting** (analyzer) and **anonymizing** (anonymizer) personal
data (PII) in text. An idiomatic Go rewrite of the core of
[Presidio](https://github.com/data-privacy-stack/presidio) (MIT), designed to be
imported as a module into a Go project — not deployed as a microservice.

```go
eng, _ := analyzer.New(analyzer.WithRegistry(registry.Default("fr")))
results, _ := eng.Analyze(ctx, "Contact : jean@exemple.fr", analyzer.MinScore(0.4))

out, _ := anonymizer.New().Anonymize(text, results, map[string]anonymizer.Operator{
    "DEFAULT": anonymizer.Replace("<PII>"),
})
```

## What PresidioGo does

The full pipeline fits into two engines:

1. **Analyzer** — scans text with *recognizers* (regex + checksum validation, plus
   optional NER) registered per language in a *registry*. Each result carries an
   entity type, a `[Start:End]` position **in runes** (safe with accents and
   emoji), a score, and an explanation. A *context boost* increases the score
   when a context word ("iban", "card", "ssn"…) precedes the entity; a
   `MinScore` threshold filters out noise.
2. **Anonymizer** — replaces detected spans according to an operator chosen per
   entity: `Replace`, `Redact`, `Mask`, `Hash`, `Keep`, `Custom`, or
   `Encrypt`/`Decrypt` (AES-GCM, with `Deanonymize` for round-trips). Overlaps
   are resolved before substitution: the entity with the highest score wins.

A regex match can be **validated** (checksum OK → maximum score), **invalidated**
(checksum KO → rejected), or left at the pattern's score — exactly Presidio's
semantics.

### Supported entities

| Group | Entities |
|---|---|
| Generic | `CREDIT_CARD` (Luhn), `EMAIL_ADDRESS`, `IBAN_CODE` (mod-97), `IP_ADDRESS` (v4/v6), `URL`, `MAC_ADDRESS`, `CRYPTO` (base58check, bech32) |
| United States | `US_SSN`, `US_ITIN`, `US_PASSPORT`, `US_DRIVER_LICENSE`, `US_BANK_NUMBER`, `ABA_ROUTING_NUMBER`, `US_NPI` (Luhn), `US_MBI`, `MEDICAL_LICENSE` (DEA) |
| France | `FR_NIR` (key 97, including Corsica), `FR_SIREN`, `FR_SIRET` (Luhn), `FR_LICENSE_PLATE` (SIV), `FR_PHONE_NUMBER` |
| NER (opt-in) | `PERSON`, `LOCATION`, `ORGANIZATION`, `NRP` via BERT-NER (ONNX) |

`registry.Default("en", "fr")` assembles generics + recognizers for each
requested language. Adding a custom recognizer means implementing a small
interface and registering it.

## Why a Go port?

Presidio is an excellent reference, but it's a **Python service**: using it from
a Go backend means deploying a container, running a Python + spaCy runtime, and
paying for an HTTP round-trip per analysis. PresidioGo flips the model:

- **A library, not a service.** `go get`, one import, zero infrastructure.
  Analysis happens in-process, with no network latency or serialization.
- **100% pure Go by default.** Without the `onnx` tag, there's no cgo dependency
  or external binary: the library cross-compiles and embeds into a single
  static binary (CLI, Lambda, sidecar…).
- **Fast and lean.** RE2 regex with guaranteed linear time, recognizers run in
  parallel (goroutines), bounded context window. Measured on a 30 KB document
  with all 28 en+fr recognizers: **13.3 ms and 3.8 MB allocated per analysis**
  (vs. 550 ms / 534 MB before optimization — ×41 in speed, ÷139 in memory).
- **Faithful to the original.** Patterns and validation rules are extracted
  from the Python code (via AST, not hand-transcribed), and an oracle corpus
  (`internal/testdata/oracle.jsonl`) drawn from Presidio's tests locks in
  non-regression. The E2E harness compares span by span against the Python
  presidio-analyzer: **100% agreement** on shared entities.
- **NER under control.** Where Presidio bundles spaCy, PresidioGo offers
  opt-in NER: a WordPiece tokenizer and BIO aggregation in pure Go, with
  BERT-NER inference via ONNX Runtime only if you enable the build tag.

## Installation

Go ≥ 1.26.

```sh
go get github.com/YoLaub/PresidioGo
```

That's it for the default usage (regex + checksums + context, pure Go).

### Quick start

```go
package main

import (
    "context"
    "fmt"

    "github.com/YoLaub/PresidioGo/analyzer"
    "github.com/YoLaub/PresidioGo/anonymizer"
    "github.com/YoLaub/PresidioGo/registry"
)

func main() {
    eng, err := analyzer.New(
        analyzer.WithRegistry(registry.Default("fr")),
        analyzer.WithDefaultLanguage("fr"),
    )
    if err != nil {
        panic(err)
    }

    text := "Email : info@presidio.site, carte 4012-8888-8888-1881, IBAN GB29NWBK60161331926819"

    results, _ := eng.Analyze(context.Background(), text, analyzer.MinScore(0.4))

    out, _ := anonymizer.New().Anonymize(text, results, map[string]anonymizer.Operator{
        "CREDIT_CARD": anonymizer.Mask('*', 15, false),
        "DEFAULT":     anonymizer.Replace("<PII>"),
    })
    fmt.Println(out.Text)
}
```

Full runnable example: `go run ./examples/basic`.

### NER (optional, `onnx` tag)

NER adds detection of people, places, and organizations via BERT-NER
(int8-quantized model, ~110 MB). It requires cgo and ONNX Runtime:

```sh
# 1. Download the model + onnxruntime 1.26.x
./scripts/download-model.sh        # or scripts/download-model.ps1 on Windows

# 2. Point to the ONNX Runtime library
export ONNXRUNTIME_LIB=/path/to/libonnxruntime.so

# 3. Build with the tag
go build -tags onnx ./...
go run -tags onnx ./examples/ner
```

Without the tag, these packages are simply absent from the build: nothing to
install, nothing to configure.

## Connection contracts

`kern-anon` is the PII brick of the [Kern](../Kern-Orch/docs/ROADMAP.md) ecosystem. Unlike
`kern-orch` and `kern-ui`, which talk over the wire, **this brick is consumed as a Go
module** — it is a library, not a service, and deliberately so: anonymisation belongs
inside the caller's process, not behind a network hop that would move PII around.

> **Naming.** The Go module is `github.com/YoLaub/PresidioGo` while the roadmap calls the
> brick `kern-anon`. The two names refer to the same thing. Renaming the module is a
> breaking change for importers and has not been done.

### Exposed — Go API

Two engines, joined by one type. Everything else is internal and may change.

```go
import (
    "github.com/YoLaub/PresidioGo/analyzer"
    "github.com/YoLaub/PresidioGo/anonymizer"
    "github.com/YoLaub/PresidioGo/registry"
)

eng, _ := analyzer.New(analyzer.WithRegistry(registry.Default("fr")))
results, _ := eng.Analyze(ctx, text, analyzer.MinScore(0.4))

out, _ := anonymizer.New().Anonymize(text, results, map[string]anonymizer.Operator{
    "DEFAULT": anonymizer.Replace("<PII>"),
})
```

| Contract | Signature | Notes |
|---|---|---|
| Detect | `analyzer.Engine.Analyze(ctx, text, ...CallOption) ([]pii.Result, error)` | Options: `Entities`, `Language`, `MinScore`. |
| Anonymise | `anonymizer.Engine.Anonymize(text, []pii.Result, map[string]Operator) (*Result, error)` | Operator key `DEFAULT` covers unlisted entity types. |
| Exchange type | `pii.Result` | Entity type, `[Start:End]` **in runes**, score, explanation. |
| Operators | `Replace`, `Mask`, `Hash`, `Redact`, `Keep`, `Encrypt`, `Decrypt`, `Custom` | |

**Guarantees a caller can rely on**

- **Offsets are rune indices, not bytes.** Safe to slice around accents and emoji.
- **No I/O, no network, no service to deploy.** The optional NER path loads a local ONNX
  model behind the `onnx` build tag; without it the library is pure Go.
- **The caller owns the text.** Nothing is logged, persisted or sent anywhere.

### Consumed

None. `kern-anon` depends on no other brick.

### Integration status

Not yet wired into any other brick. The roadmap marks it `externe · fait · integration to do`
— `kern-orch` does not call it today, and no contract exists yet for where in a run PII
scrubbing would happen.

## Development

```sh
go test ./...          # tests (TDD, oracle corpus)
golangci-lint run      # lint
go run ./internal/oracleharness   # E2E vs Python presidio-analyzer (Docker, port 5002)
```

Status: v0.2 — full pipeline (analyzer, anonymizer, 21 en/fr recognizers,
ONNX NER, oracle harness). Roadmap: `docs/PLAN.md`, per-feature notes:
`docs/index/`.

## License

MIT — derived from the architecture of Microsoft Presidio (MIT).
