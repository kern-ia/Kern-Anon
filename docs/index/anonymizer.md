---
id: okf-004
feature: anonymizer
branch: feature/anonymizer
status: done
files:
  - anonymizer/operators.go (Replace, Redact, Keep, Custom, Mask, Hash, Encrypt/Decrypt AES-GCM)
  - anonymizer/engine.go (Engine.Anonymize/Deanonymize, resolveConflicts, Item/Result)
tests:
  - anonymizer/anonymizer_test.go (12 tests, 90.9 % — chevauchements, runes, round-trip)
decisions:
  - "2026-07-16 : operators = closures derrière une interface 2 méthodes (pas de factory — idiomatique Go, remplace OperatorsFactory)"
  - "2026-07-16 : chiffrement AES-GCM base64(nonce||ct) au lieu de l'AES-CBC du fork (authentifié, round-trip Go only)"
  - "2026-07-16 : chevauchements = score desc, longueur desc, position asc, sélection gloutonne sans intersection"
  - "2026-07-16 : Item expose les offsets (runes) dans le texte DE SORTIE — nécessaires au Deanonymize"
---

**Quoi** : le moteur d'anonymisation — opérateur choisi par entité puis DEFAULT
puis Replace("<ENTITY_TYPE>") ; résolution des chevauchements avant substitution ;
Deanonymize rejoue les items (typiquement Decrypt) pour le round-trip.

**Pièges** : rien de nouveau — le seul rouge du TDD était une erreur d'arithmétique
dans un attendu de test (16−12=4 chiffres restants, pas 6).
