---
id: okf-005
feature: context-enhancer
branch: feature/context-enhancer
status: done
files:
  - contextaware/enhancer.go (Enhancer, défauts du fork 0.35/0.4/5-avant/0-après)
  - recognizer/pattern_recognizer.go (WithContextWords, ContextWords())
  - recognizers/generic/*.go (mots de contexte CONTEXT du fork ajoutés aux 7)
tests:
  - contextaware/enhancer_test.go (7 tests : boost, fenêtre, min, plafond)
decisions:
  - "2026-07-16 : matching « substring » sur la fenêtre jointe en minuscules (mode du fork) — sans lemmatisation tant que le NLP n'est pas branché"
  - "2026-07-16 : recognizer d'origine retrouvé via Explanation.Recognizer ; interface ContextAware par assertion de type"
  - "2026-07-16 : mots de contexte injectés par script Python (7 fichiers d'un coup, assertions de comptage)"
---

**Quoi** : boost contextuel — si un mot de contexte du recognizer apparaît dans
les 5 mots précédant l'entité, score +0.35 (min 0.4, plafonné à 1.0),
ContextBoost tracé dans l'Explanation.

**Pièges** :
- Le matching substring boostera « ip » dans « description » : fidèle au fork,
  à revisiter quand la lemmatisation NLP sera branchée.
