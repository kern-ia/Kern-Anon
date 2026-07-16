---
id: okf-006
feature: analyzer
branch: feature/analyzer
status: done
files:
  - analyzer/analyzer.go (Engine, options construction + options d'appel)
  - nlp/nlp.go (interface Engine + NoOp — prêt pour nlp-onnx)
  - examples/basic/main.go (E2E : API publique du plan, pipeline complet)
tests:
  - analyzer/analyzer_test.go (8 tests, 88.4 % — pipeline, filtres, dédoublonnage, câblage NLP)
decisions:
  - "2026-07-16 : options d'appel fonctionnelles (Language/Entities/MinScore) — API du plan §4 réalisée telle quelle"
  - "2026-07-16 : dédoublonnage = span contenu dans un span du MÊME type au score ≥ ; les chevauchements inter-entités restent (comme le fork) et sont tranchés par l'anonymizer"
  - "2026-07-16 : enhancer contextuel activé par défaut, WithEnhancer(nil) pour le couper"
  - "2026-07-16 : nlpEngine.Load() appelé dans New — échec de chargement = erreur de construction, pas d'appel"
---

**Quoi** : le moteur qui orchestre le pipeline — NLP (optionnel) → recognizers
du registry → boost contextuel → dédoublonnage même-entité → seuil → tri par
position. C'est l'API d'entrée de la lib.

**Pièges** : aucun nouveau — le câblage NLP est vérifié par un fake recognizer
qui capture les artifacts.
