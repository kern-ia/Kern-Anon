---
id: okf-013
feature: ner-sliding-window
branch: feature/ner-sliding-window
status: done
files:
  - nlp/bertner/window.go (Windows + MergeEntities)
  - nlp/bertner/tokenizer.go (Aggregate : jamais de coupure intra-mot)
  - nlp/onnx/engine.go (fenêtres 256/chevauchement 64, inférences parallèles, RWMutex)
  - examples/ner/main.go (E2E texte long, entités en fin de document)
tests:
  - nlp/bertner/window_test.go + tokenizer_test.go (92.5 %)
decisions:
  - "2026-07-17 : fenêtres de 256 tokens (pas 512) — les labels se dégradent en fin de séquence longue ET l'attention O(n²) rend 2×256 moins cher qu'1×512"
  - "2026-07-17 : chevauchement 64 tokens ; fusion MergeEntities = span le plus long puis score"
  - "2026-07-17 : sessions ORT thread-safe pour Run → RWMutex (Load/Destroy en écriture), fenêtres inférées en parallèle"
  - "2026-07-17 : une entité ne se termine jamais au milieu d'un mot (sous-mot ## étiqueté O rattaché)"
---

**Quoi** : le NER voit désormais TOUT le texte — fenêtres chevauchantes
inférées en parallèle, fusion des doublons inter-fenêtres, mots jamais coupés.

**Mesure/E2E** : entités en position 4700+ (600+ tokens) : avant = invisibles ;
après = « Marie Curie » PERSON 0.96, « University of Warsaw » ORG 0.95.

**Pièges** :
- En fenêtre 512 pleine, les labels du modèle se dégradent près de la fin de
  séquence (« Marie » → ORG 0.64) : symptôme modèle, pas code — d'où les
  fenêtres 256.
- Le dernier sous-mot d'un nom peut être étiqueté O par le modèle
  (« Curi/##e ») : rattacher les continuations ## à l'entité ouverte.
