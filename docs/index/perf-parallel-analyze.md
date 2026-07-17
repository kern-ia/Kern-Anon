---
id: okf-012
feature: perf-parallel-analyze
branch: feature/perf-parallel-analyze
status: done
files:
  - contextaware/enhancer.go (fenêtre par balayage arrière/avant borné — plus de découpage du préfixe entier)
  - analyzer/analyzer.go (fan-out goroutines des recognizers, réassemblage dans l'ordre du registry)
tests:
  - analyzer/analyzer_bench_test.go (document 30 Ko, registry en+fr complet)
  - suites analyzer + contextaware inchangées, passées avec -race
decisions:
  - "2026-07-17 : le benchmark a montré que le VRAI goulot était l'enhancer (O(résultats × longueur du texte)), pas la séquentialité — toujours mesurer avant d'optimiser"
  - "2026-07-17 : déterminisme du fan-out garanti par réassemblage indexé dans l'ordre du registry"
---

**Quoi** : les deux optimisations du pipeline — fenêtre contextuelle en
balayage borné au voisinage de l'entité, et exécution parallèle des
recognizers (indépendants et thread-safe).

**Mesure** (document 30 Ko, 28 recognizers) : 550 ms/op → 13.3 ms/op (×41),
534 Mo → 3.8 Mo alloués par analyse (÷139).
