---
id: okf-008
feature: recognizers-fr
branch: feature/recognizers-fr
status: done
files:
  - recognizers/fr/fr.go (5 recognizers created — no FR source in the fork)
  - registry/default.go (registry.Default(languages...) — the plan's API)
  - internal/testdata/oracle.jsonl (+10 FR cases → 35)
tests:
  - recognizers/fr/fr_test.go (oracle + NirKey incl. Corsica 2A/2B, 93.1%)
  - registry/registry_test.go (Default: 16 generic recognizers, 12 fr)
decisions:
  - "2026-07-17: FR_NIR validated by key 97 (2A→19, 2B→18) — SYNTHETIC test values computed, never real NIRs"
  - "2026-07-17: SIREN/SIRET validated with Luhn — documented limit: La Poste SIRENs (356000000) don't pass Luhn"
  - "2026-07-17: SIV plate AA-123-AA with I/O/U excluded; national + international phone (+33/0033)"
  - "2026-07-17: registry.Default lives in registry (no cycle: registry→recognizers/*→recognizer)"
---

**What**: the French locale — FR_NIR (key 97), FR_SIREN, FR_SIRET (Luhn),
FR_LICENSE_PLATE (SIV), FR_PHONE_NUMBER — plus registry.Default which
assembles generics + per-language locales.

**Pitfalls**:
- Windows cp1252 console: the Python computation scripts need
  PYTHONIOENCODING=utf-8 set (otherwise UnicodeEncodeError on "→").
