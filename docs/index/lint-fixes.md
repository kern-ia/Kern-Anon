---
id: okf-014
feature: lint-fixes
branch: feature/lint-fixes
status: done
files:
  - .golangci.yml (misspell désactivé — commentaires en français)
  - internal/oracleharness, internal/oracletest (errcheck sur Close)
  - recognizer, recognizers/ner (doc des méthodes exportées)
  - nlp/bertner/tokenizer.go (prealloc), recognizers/generic/creditcard.go (De Morgan)
  - gofmt -w sur le repo
tests:
  - golangci-lint run : 0 issue (même config que la CI) ; suite -race verte (10 pkgs)
decisions:
  - "2026-07-17 : misspell retiré des linters — il suppose l'anglais et corrige le français (« contextuel »→« contextual ») ; le reste des findings était légitime et corrigé"
  - "2026-07-17 : golangci-lint v2 installé en local (go install …/v2/cmd@latest) pour reproduire la CI avant de pousser"
---

**Quoi** : mise au vert du job Lint de la CI GitHub — 28 findings corrigés
(errcheck, gofmt, revive exported/unused-parameter, prealloc, staticcheck)
et misspell écarté car incompatible avec des commentaires français.

**Pièges** : reproduire le lint en local AVANT de pousser — la CI n'était que
le messager.
