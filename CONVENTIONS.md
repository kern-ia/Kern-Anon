# CONVENTIONS.md — kern-anon

Autorité locale pour ce repo, comme annoncé par le [CONTRIBUTING.md](https://github.com/kern-ia/.github/blob/main/CONTRIBUTING.md)
de l'organisation. Les règles communes à tous les repos `kern-ia` sont reprises ci-dessous ;
la section « Spécificités » couvre ce qui n'appartient qu'à `kern-anon`.

## Branches

- `main` : branche stable, toujours déployable. Protégée — aucun push direct.
- `dev` : branche d'intégration. Protégée — aucun push direct.
- Branches de travail : `feature/<slug>`, `fix/<slug>`, `chore/<slug>`, `docs/<slug>`, `test/<slug>`.
- Toute modification de `main` ou `dev` passe par une Pull Request. Le `CLAUDE.md` de ce repo
  dit déjà « jamais de commit direct sur main/dev » — cette règle n'est aujourd'hui pas
  appliquée techniquement (aucune protection de branche activée sur GitHub) et pas non plus
  suivie via PR (0 PR ouverte à ce jour, tout part de merges locaux poussés directement).
  À corriger : activer la protection de branche ET passer par PR même en solo.
- Merge vers `dev` : `--no-ff` UNIQUEMENT si tests verts + E2E fait (règle déjà en vigueur,
  à garder).

## Commits

Conventional Commits : `feat:`, `fix:`, `test:`, `docs:`, `chore:`… (déjà respecté).
Aucune signature d'outil (trailer `Co-Authored-By`, `Claude-Session` ou équivalent) dans les
messages de commit — l'auteur du commit git suffit.

## Pull Requests

- Un seul sujet par PR, liée à l'issue ou la RFC qu'elle résout.
- Template PR hérité de `kern-ia/.github`.
- Déclare l'impact semver.
- Aucune donnée personnelle réelle — critique ici puisque le repo manipule justement de la PII
  synthétique par construction.

## Méthode (déjà en vigueur, à garder)

- **TDD strict** : tests écrits avant le code, `go test ./...` vert avant tout commit.
- Cas de test dérivés du corpus oracle (`internal/testdata/oracle.jsonl` / `anonymize.jsonl`).
- Logique métier pure dans les packages domaine (`recognizer`, `analyzer`, `anonymizer`…),
  aucune I/O dedans.
- Offsets exprimés en **runes**, jamais en bytes — testé avec accents/emoji sur chaque
  recognizer.
- `nlp/onnx` reste opt-in derrière le build tag `onnx` (cgo) ; sans le tag, 100 % Go pur.

## Style et lint

`.golangci.yml` — `version: 2`, `linters.default: standard` + `revive`, `gocritic`, `prealloc`.
`misspell` désactivé délibérément (commentaires en français). Toute extension future du set de
linters doit porter le même commentaire justificatif que l'existant. `max-issues-per-linter: 0`,
`max-same-issues: 0` — rien n'est masqué.

## Tests / CI

- `go build ./...`, `go build -tags onnx ./...`, `go test -race -cover ./...`, lint
  `golangci-lint`. Déjà en place dans `.github/workflows/ci.yml` — à garder comme référence
  pour les autres repos Go de l'org qui n'ont pas encore de CI (`kern-orch`).

## Module Go

> **Écart actuel — le plus visible de l'audit** : `go.mod` déclare encore
> `module github.com/YoLaub/PresidioGo`, reliquat du nom d'avant le renommage en `kern-anon`.
> Ce chemin ne correspond ni au nom du repo (`kern-anon`) ni à l'organisation (`kern-ia`).
> À corriger : renommer le module (`gofmt -r`/`go mod edit -module` + mise à jour de tous les
> imports internes), en coordination avec la décision d'organisation sur le chemin
> `github.com/kern-ia/...` (voir rapport global).

## Documentation

- `README.md`, `LICENSE` à la racine (les deux présents — bon exemple pour les autres repos).
- `CLAUDE.md` — contexte agent.
- Index OKF sous `docs/index/` : une fiche par feature terminée (entête YAML : id, feature,
  branch, status, files, tests, decisions ; corps ≤ 15 lignes). Pièges rencontrés consignés
  dans `docs/index/retro.md` au moment où ils surviennent.
- Pas de `CHANGELOG.md` : notes de version dans le tag annoté (convention org).

## Sécurité / confidentialité

Voir `SECURITY.md` hérité de l'org. Les défauts de frontière de confidentialité (PII non
pseudonymisée qui atteint un provider, contenu verbatim loggé) sont traités avec la même
sévérité qu'une faille d'exécution de code — rappel explicite pour ce repo en particulier.
