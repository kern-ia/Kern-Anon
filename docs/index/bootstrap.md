---
id: okf-001
feature: bootstrap
branch: feature/bootstrap
status: done
files:
  - go.mod (module github.com/YoLaub/PresidioGo, go 1.26)
  - CLAUDE.md, README.md, LICENSE (MIT), .gitignore, .golangci.yml
  - .github/workflows/ci.yml (build + test -race + lint)
  - docs/PLAN.md (plan validated 2026-07-16)
  - internal/testdata/oracle.jsonl (seed corpus, 8 cases)
tests:
  - internal/testdata/oracle.jsonl (consumed by the following features)
decisions:
  - "2026-07-16: go 1.26 (toolchain installed via winget, go1.26.5)"
  - "2026-07-16: rune-based offsets across the whole API (compat with the Python corpus)"
  - "2026-07-16: ONNX NER behind the `onnx` build tag — default is 100% pure Go"
  - "2026-07-16: golangci-lint v2 config (default: standard + revive/gocritic)"
---

**What**: repo skeleton — Go module, lint/CI tooling, conventions
(CLAUDE.md), validated plan, seed oracle corpus extracted from the
presidigo fork's Python tests.

**Pitfalls**:
- Go missing from the machine: `winget install GoLang.Go`, then refresh the
  session PATH before `go version`.
