---
id: okf-007
feature: recognizers-us
branch: feature/recognizers-us
status: done
files:
  - recognizers/us/us.go (9 recognizers, exported ABA/NPI/DEA checksums)
  - internal/oracletest/oracletest.go (shared oracle runner, filtered by supported entities)
  - internal/testdata/oracle.jsonl (+8 US cases → 25)
tests:
  - recognizers/us/us_test.go (oracle + checksum unit tests, 86%)
decisions:
  - "2026-07-17: patterns extracted from the fork via Python AST (no manual retranscription)"
  - "2026-07-17: the fork's driver-license bug kept as-is (\"A-Z]{2}\" missing a bracket) for iso-behavior — documented in the code"
  - "2026-07-17: NPI = Luhn(80840+npi) + rejection of degenerate bodies (fork's invalidate)"
  - "2026-07-17: oracle runner factored into internal/oracletest — each locale only evaluates its own entities"
---

**What**: the fork's 9 US recognizers — US_SSN, US_ITIN, US_PASSPORT,
US_DRIVER_LICENSE, US_BANK_NUMBER, ABA_ROUTING_NUMBER (3-7-1 weighted sum),
US_NPI (Luhn prefixed with 80840), US_MBI (CMS charset), MEDICAL_LICENSE
(DEA).

**Pitfalls**:
- No non-RE2 regex on the US side: direct port.
- The fork's DEA validation skips the 2 letters then weights even positions
  ×2 — verified by manual calculation (AB1234563 → 33 → 3).
