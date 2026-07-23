---
id: okf-015
feature: anonymize-e2e-demo
branch: main
status: done
files:
  - internal/testdata/anonymize.jsonl (6 cases fr/en, full pipeline golden output)
  - internal/anonymizeoracle/anonymizeoracle.go (loader + Run: analyzer → anonymizer, compares expected_text)
  - anonymizer/anonymizer_test.go (TestAnonymize_OracleE2E wires the corpus in)
  - examples/anonymize-demo/main.go (visual demo: Mask/Hash/Redact/Replace/Encrypt+Decrypt round-trip, fr+en)
tests:
  - anonymizer/anonymizer_test.go::TestAnonymize_OracleE2E (6/6 green)
decisions:
  - "2026-07-23: golden corpus uses the default nil-operator map (Replace <ENTITY_TYPE>) so expected_text is fully deterministic from what the analyzer detects — also exercises overlap resolution (URL spans nested in EMAIL_ADDRESS are dropped)"
  - "2026-07-23: separate from oracle.jsonl/oracletest (recognizer-level) — this corpus locks the END-TO-END behavior (analyzer + registry.Default + anonymizer together)"
---

**What**: a demonstration test set proving the anonymizer works correctly
end-to-end, not just at the unit level. Two parts: a golden-file corpus
(`internal/testdata/anonymize.jsonl`) replayed by a Go test
(`TestAnonymize_OracleE2E`) for non-regression, and a runnable example
(`examples/anonymize-demo`) that prints detected entities, the anonymized
text under several operators, and a decrypt round-trip — for visual review.

**Pitfalls**: none — reused already-validated synthetic PII values from
`oracle.jsonl` (NIR, SIREN, SSN) to avoid inventing new checksums by hand.
