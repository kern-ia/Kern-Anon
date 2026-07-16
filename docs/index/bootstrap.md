---
id: okf-001
feature: bootstrap
branch: feature/bootstrap
status: done
files:
  - go.mod (module github.com/YoLaub/presidigo-go, go 1.26)
  - CLAUDE.md, README.md, LICENSE (MIT), .gitignore, .golangci.yml
  - .github/workflows/ci.yml (build + test -race + lint)
  - docs/PLAN.md (plan validé le 2026-07-16)
  - internal/testdata/oracle.jsonl (corpus seed, 8 cas)
tests:
  - internal/testdata/oracle.jsonl (consommé par les features suivantes)
decisions:
  - "2026-07-16 : go 1.26 (toolchain installée via winget, go1.26.5)"
  - "2026-07-16 : offsets en runes dans toute l'API (compat corpus Python)"
  - "2026-07-16 : NER ONNX derrière build tag `onnx` — défaut 100 % Go pur"
  - "2026-07-16 : golangci-lint config v2 (default: standard + revive/gocritic)"
---

**Quoi** : squelette du repo — module Go, outillage lint/CI, conventions (CLAUDE.md),
plan validé, corpus oracle seed extrait des tests Python du fork presidigo.

**Pièges** :
- Go absent de la machine : `winget install GoLang.Go`, puis rafraîchir le PATH
  de la session avant `go version`.
