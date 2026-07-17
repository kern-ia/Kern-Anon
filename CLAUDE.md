# CLAUDE.md — Conventions du repo presidigo-go

## Contexte
Bibliothèque Go de détection (**analyzer**) et d'anonymisation (**anonymizer**) de PII
dans du texte. Refonte Go du cœur de Presidio, destinée à être importée comme module
(`github.com/YoLaub/PresidioGo`) dans un projet Go. Périmètre v1 : texte seul.
Plan de référence : `docs/PLAN.md`. Périmètre fonctionnel d'origine :
`../presidigo/archi-output/INDEX.md`.

## Contexte de travail : index OKF d'abord
- Lire `docs/index/` en début de session au lieu de relire tout le code.
- Chaque feature terminée = fiche `docs/index/<feature>.md` (entête YAML : id, feature,
  branch, status, files, tests, decisions ; corps ≤ 15 lignes).
- Pièges rencontrés → `docs/index/retro.md` AU MOMENT où ils mordent.

## Méthode obligatoire
- **TDD strict** : tests écrits AVANT le code (`go test ./...` vert avant tout commit).
  Les cas de test viennent du corpus oracle (`internal/testdata/oracle.jsonl`),
  extrait des tests Python du fork presidigo.
- **Architecture** : logique pure dans les packages métier (`recognizer`, `analyzer`,
  `anonymizer`…) ; aucune I/O dans ces packages. Interfaces petites et idiomatiques.
- **Offsets en runes** : toute position (Start/End) est exprimée en runes, pas en bytes —
  testé avec accents/emoji dans chaque recognizer.
- **NER opt-in** : le package `nlp/onnx` est derrière le build tag `onnx` (cgo).
  Sans le tag, la lib est 100 % Go pur avec `nlp.NoOp`.
- **Git** : `main` ← `dev` ← `feature/xx`. Jamais de commit direct sur main/dev.
  Merge `--no-ff` vers dev UNIQUEMENT si tests verts + E2E fait.
  Commits conventionnels (`feat:`, `fix:`, `test:`, `docs:`…).

## Commandes
- Tests : `go test ./...` — lint : `golangci-lint run` — build : `go build ./...`
- Sans cgo/ONNX (défaut) : rien à faire. Avec NER : `go build -tags onnx ./...`
- E2E oracle vs Python : `go run ./internal/oracleharness` (nécessite le
  presidio-analyzer du fork : `docker compose up -d` dans ../presidigo, port 5002)

## Règles de sécurité
- **INTERDIT : toute suppression de fichiers ou dossiers (`rm`, `rm -rf`,
  `Remove-Item`, `del`, `git clean`…) sans autorisation explicite de l'utilisateur
  au préalable.**
- Jamais de PII réelles dans les tests ou le corpus — uniquement des données
  synthétiques (les numéros valides au checksum sont générés, pas collectés).

## Règles métier clés
- Un match regex peut être **validé** (checksum OK → score max), **invalidé**
  (checksum KO → résultat rejeté) ou laissé au score du pattern.
- Chevauchements à l'anonymisation : les spans sont triés puis fusionnés ;
  l'entité au score le plus haut gagne.
- Le boost contextuel (`contextaware`) ne s'applique qu'aux résultats des
  recognizers qui déclarent des mots de contexte.
