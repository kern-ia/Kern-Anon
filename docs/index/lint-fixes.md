---
id: okf-014
feature: lint-fixes
branch: feature/lint-fixes
status: done
files:
  - .golangci.yml (misspell disabled — comments are in French)
  - internal/oracleharness, internal/oracletest (errcheck on Close)
  - recognizer, recognizers/ner (doc comments on exported methods)
  - nlp/bertner/tokenizer.go (prealloc), recognizers/generic/creditcard.go (De Morgan)
  - gofmt -w on the repo
tests:
  - golangci-lint run: 0 issues (same config as CI); -race suite green (10 pkgs)
decisions:
  - "2026-07-17: misspell removed from the linters — it assumes English and \"corrects\" French words (\"contextuel\"→\"contextual\"); the rest of the findings were legitimate and fixed"
  - "2026-07-17: golangci-lint v2 installed locally (go install …/v2/cmd@latest) to reproduce CI before pushing"
---

**What**: greening the CI Lint job — 28 findings fixed (errcheck, gofmt,
revive exported/unused-parameter, prealloc, staticcheck) and misspell
dropped because it's incompatible with French-language comments.

**Pitfalls**: reproduce lint locally BEFORE pushing — CI was just the
messenger.
