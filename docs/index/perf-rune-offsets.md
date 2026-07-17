---
id: okf-011
feature: perf-rune-offsets
branch: feature/perf-rune-offsets
status: done
files:
  - recognizer/pattern_recognizer.go (collecte bytes puis runeOffsets en une passe)
tests:
  - recognizer/pattern_recognizer_bench_test.go (60 Ko accentué, 800 matches)
  - suite existante inchangée (offsets identiques)
decisions:
  - "2026-07-17 : conversion en une passe via `for byteIdx := range text` (itère les débuts de rune) — les frontières regex sont toujours des frontières de rune (RE2)"
  - "2026-07-17 : benchmark AVANT optimisation pour chiffrer la baseline (25.3 ms/op)"
---

**Quoi** : suppression du comportement quadratique de la conversion
bytes→runes — `utf8.RuneCountInString` par match remplacé par une passe
unique sur le texte pour tous les matches.

**Mesure** : 25.3 ms/op → 2.39 ms/op (×10.6), 2.6 → 27.9 Mo/s, même texte
(60 Ko, 800 matches). Aucun changement de comportement (suite verte).
