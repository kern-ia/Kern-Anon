---
id: okf-010
feature: oracle-harness
branch: feature/oracle-harness
status: done
files:
  - internal/oracleharness/main.go (Go vs Python presidio-analyzer comparison)
  - recognizers/us/us.go (ssnValidate — fork's invalidate_result ported)
  - internal/testdata/oracle.jsonl (SSN: valid case 216-09-1234 + 2 negatives → 37 cases)
tests:
  - harness run against the fork's Docker container: 16/16 (100%)
decisions:
  - "2026-07-17: comparison restricted to entities in BOTH registries — the default Python service doesn't load ABA/NPI/MBI (verified via /recognizers)"
  - "2026-07-17: NER, DATE_TIME and PHONE_NUMBER excluded (different engines/recognizers)"
  - "2026-07-17: v0.1.0 criterion (≥95%) MET: 100% agreement"
---

**What**: the E2E harness from plan §7 — runs the Go engine on the oracle
corpus and compares span by span against the Python presidio-analyzer's
POST /analyze (fork's container). Output: list of divergences + agreement
percentage.

**Pitfalls** (the harness paid off immediately):
- The fork INVALIDATES suspicious SSNs (mixed delimiters, all-zero groups,
  000/666 area, published canonical SSNs including 078-05-1120) — rules not
  ported on the first pass, caught by the harness, ported into ssnValidate.
- The Python service's default registry ≠ the code's class list: always
  check /recognizers before comparing.
