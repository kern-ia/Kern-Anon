---
id: okf-008
feature: recognizers-fr
branch: feature/recognizers-fr
status: done
files:
  - recognizers/fr/fr.go (5 recognizers créés — pas de source FR dans le fork)
  - registry/default.go (registry.Default(langues...) — l'API du plan)
  - internal/testdata/oracle.jsonl (+10 cas FR → 35)
tests:
  - recognizers/fr/fr_test.go (oracle + NirKey dont Corse 2A/2B, 93.1 %)
  - registry/registry_test.go (Default : 16 recognizers en, 12 fr)
decisions:
  - "2026-07-17 : FR_NIR validé par clé 97 (2A→19, 2B→18) — valeurs de test SYNTHÉTIQUES calculées, jamais de vrais NIR"
  - "2026-07-17 : SIREN/SIRET validés Luhn — limite documentée : les SIREN La Poste (356000000) ne passent pas Luhn"
  - "2026-07-17 : plaque SIV AA-123-AA avec I/O/U exclues ; téléphone national + international (+33/0033)"
  - "2026-07-17 : registry.Default vit dans registry (pas de cycle : registry→recognizers/*→recognizer)"
---

**Quoi** : la locale française — FR_NIR (clé 97), FR_SIREN, FR_SIRET (Luhn),
FR_LICENSE_PLATE (SIV), FR_PHONE_NUMBER — plus registry.Default qui assemble
génériques + locales par langue.

**Pièges** :
- Console Windows cp1252 : les scripts Python de calcul doivent poser
  PYTHONIOENCODING=utf-8 (sinon UnicodeEncodeError sur « → »).
