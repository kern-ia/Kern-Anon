---
id: okf-002
feature: core-types
branch: feature/core-types
status: done
files:
  - pii/types.go (Result, Pattern, Explanation, MaxScore)
  - nlp/nlp.go (Artifacts minimal — l'interface moteur viendra avec nlp-onnx)
  - recognizer/recognizer.go (interface Recognizer, ValidateFunc)
  - recognizer/pattern_recognizer.go (NewPattern, WithValidate, offsets runes)
  - registry/registry.go (Add/Remove/Get(langue, entités...)/SupportedEntities, RWMutex)
  - examples/basic/main.go (E2E réel)
tests:
  - recognizer/pattern_recognizer_test.go (100 % couverture)
  - registry/registry_test.go (100 % couverture)
decisions:
  - "2026-07-16 : ValidateFunc retourne *bool — nil neutre, true→MaxScore, false→rejet (calqué sur validate/invalidate_result du fork)"
  - "2026-07-16 : conversion bytes→runes via utf8.RuneCountInString au point de match (pas de table précalculée tant que non mesuré)"
  - "2026-07-16 : Registry protégé par RWMutex — setup mutable, Analyze concurrent"
---

**Quoi** : le socle — types partagés, interface Recognizer, PatternRecognizer
(regex + validation checksum), Registry filtrant par langue/entité. Offsets runes
vérifiés sur texte accentué (cas oracle email-rune-offsets).

**Pièges** :
- `go test -race` exige cgo/gcc, absent sous Windows par défaut → -race délégué à la CI Linux.
