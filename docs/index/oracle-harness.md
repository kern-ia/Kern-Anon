---
id: okf-010
feature: oracle-harness
branch: feature/oracle-harness
status: done
files:
  - internal/oracleharness/main.go (comparaison Go vs presidio-analyzer Python)
  - recognizers/us/us.go (ssnValidate — invalidate_result du fork porté)
  - internal/testdata/oracle.jsonl (SSN : cas valide 216-09-1234 + 2 négatifs → 37 cas)
tests:
  - harness exécuté contre le conteneur Docker du fork : 16/16 (100 %)
decisions:
  - "2026-07-17 : comparaison restreinte aux entités des DEUX registres — le service Python par défaut ne charge pas ABA/NPI/MBI (vérifié via /recognizers)"
  - "2026-07-17 : NER, DATE_TIME et PHONE_NUMBER exclus (moteurs/recognizers différents)"
  - "2026-07-17 : critère v0.1.0 (≥95 %) ATTEINT : 100 % d'accord"
---

**Quoi** : le harness E2E du plan §7 — lance le moteur Go sur le corpus oracle
et compare span par span avec le POST /analyze du presidio-analyzer Python
(conteneur du fork). Sortie : liste des divergences + pourcentage d'accord.

**Pièges** (le harness a payé immédiatement) :
- Le fork INVALIDE les SSN suspects (delimiteurs mélangés, groupes à zéros,
  zone 000/666, SSN canoniques publiés dont 078-05-1120) — règles non portées
  au premier passage, détectées par le harness, portées dans ssnValidate.
- Le registre par défaut du service Python ≠ la liste des classes du code :
  toujours vérifier /recognizers avant de comparer.
