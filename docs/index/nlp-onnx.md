---
id: okf-009
feature: nlp-onnx
branch: feature/nlp-onnx
status: done
files:
  - nlp/bertner/tokenizer.go (WordPiece pur Go, offsets runes + Aggregate BIO)
  - nlp/onnx/engine.go (nlp.Engine via onnxruntime, build tag onnx)
  - recognizers/ner/ner.go (artifacts → PII : PER→PERSON, LOC→LOCATION, ORG→ORGANIZATION, MISC→NRP)
  - nlp/nlp.go (NerEntity, Artifacts.NerEntities)
  - scripts/download-model.{ps1,sh} (Xenova/bert-base-NER quantisé + onnxruntime 1.26.0)
  - examples/ner/main.go (E2E tag onnx)
tests:
  - nlp/bertner/tokenizer_test.go (90.9 % — vocab embarqué, runes, BIO)
  - recognizers/ner/ner_test.go (95 % — mapping, artifacts nil, label inconnu)
decisions:
  - "2026-07-17 : modèle = Xenova/bert-base-NER quantisé int8 (~110 Mo) — qualité bert-base, taille contenue"
  - "2026-07-17 : onnxruntime chargé dynamiquement (ONNXRUNTIME_LIB) — version 1.26.x requise par yalue/onnxruntime_go v1.31 (API 26)"
  - "2026-07-17 : tokenizer CASED par défaut (do_lower_case lu de tokenizer_config.json) — le lowercase détruisait le NER"
  - "2026-07-17 : gcc local = winlibs (winget BrechtSanders.WinLibs.POSIX.UCRT) → cgo et -race disponibles sous Windows"
---

**Quoi** : le NER — tokenizer WordPiece et agrégation BIO en pur Go (testables
partout), inférence BERT-NER ONNX derrière le tag `onnx`, recognizer ner qui
mappe les entités vers PERSON/LOCATION/ORGANIZATION/NRP. E2E local vérifié :
« John Smith »/« Microsoft »/« Seattle » détectés à 1.00.

**Pièges** :
- Version onnxruntime ↔ yalue : l'API doit correspondre (v1.31 → ORT 1.26.x),
  sinon « API version [26] is not available ».
- Modèle cased : tokeniser en minuscules fait disparaître toutes les entités.
- .ps1 accentué sans BOM → parse error PowerShell 5.1 (ANSI) : scripts en ASCII.
