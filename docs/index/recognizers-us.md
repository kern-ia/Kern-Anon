---
id: okf-007
feature: recognizers-us
branch: feature/recognizers-us
status: done
files:
  - recognizers/us/us.go (9 recognizers, checksums ABA/NPI/DEA exportés)
  - internal/oracletest/oracletest.go (runner oracle partagé, filtré par entités supportées)
  - internal/testdata/oracle.jsonl (+8 cas US → 25)
tests:
  - recognizers/us/us_test.go (oracle + unitaires checksums, 86 %)
decisions:
  - "2026-07-17 : patterns extraits du fork par AST Python (pas de retranscription manuelle)"
  - "2026-07-17 : bug du fork conservé tel quel dans le permis (« A-Z]{2} » sans crochet) pour l'iso-comportement — documenté dans le code"
  - "2026-07-17 : NPI = Luhn(80840+npi) + rejet des corps dégénérés (invalidate du fork)"
  - "2026-07-17 : runner oracle factorisé dans internal/oracletest — chaque locale n'évalue que ses entités"
---

**Quoi** : les 9 recognizers US du fork — US_SSN, US_ITIN, US_PASSPORT,
US_DRIVER_LICENSE, US_BANK_NUMBER, ABA_ROUTING_NUMBER (somme 3-7-1),
US_NPI (Luhn préfixé 80840), US_MBI (charset CMS), MEDICAL_LICENSE (DEA).

**Pièges** :
- Aucune regex non-RE2 côté US : portage direct.
- La validation DEA du fork saute les 2 lettres puis pondère ×2 les positions
  paires — vérifiée par calcul manuel (AB1234563 → 33 → 3).
