---
id: okf-003
feature: generic-recognizers
branch: feature/generic-recognizers
status: done
files:
  - recognizers/generic/generic.go (All(language), mustPattern)
  - recognizers/generic/checksums.go (Luhn, IbanMod97)
  - recognizers/generic/creditcard.go, email.go, iban.go, ip.go, url.go, mac.go, crypto.go
  - internal/testdata/oracle.jsonl (17 cas, positifs ET négatifs)
tests:
  - recognizers/generic/generic_test.go (runner oracle + unitaires checksums, 90.8 %)
decisions:
  - "2026-07-16 : lookahead carte (exclusion 1+12 chiffres) porté en validation Go (RE2 sans lookaround)"
  - "2026-07-16 : backreference MAC \\1 scindée en patterns par séparateur (sémantique préservée)"
  - "2026-07-16 : IPv6 = regex candidate large + validation net/netip (plus fiable que la regex exhaustive du fork)"
  - "2026-07-16 : IBAN générique + mod-97 (les patterns par pays du fork viendront si besoin)"
  - "2026-07-16 : phone (dép. libphonenumber/nyaruka) et date reportés à une feature dédiée"
  - "2026-07-16 : url.go GÉNÉRÉ par extraction sed depuis url_recognizer.py — ne pas éditer la regex à la main"
---

**Quoi** : les 7 recognizers génériques portés du fork — CREDIT_CARD (Luhn),
EMAIL_ADDRESS, IBAN_CODE (mod-97), IP_ADDRESS (netip), URL, MAC_ADDRESS,
CRYPTO (base58check + bech32/bech32m). Corpus oracle porté à 17 cas.

**Pièges** :
- RE2 : lookahead/lookbehind/backreferences des regex Python à réécrire (3 cas ici).
- Le score email reste 0.5 (pas de validation TLD tldextract en v1).
