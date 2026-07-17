# Rétro continue — presidigo-go

## 2026-07-17 — Rétro v0.2.0 (performance)
**Leçon majeure** : le benchmark baseline a invalidé l'intuition — le goulot
n'était pas la séquentialité des recognizers mais la fenêtre contextuelle
(O(résultats × texte)) : 550 ms et 534 Mo alloués sur 30 Ko. Toujours mesurer
AVANT d'optimiser. Gains livrés : recognizer ×10,6 (offsets runes en une
passe), pipeline ×41 (fenêtre bornée + fan-out), NER fenêtré 256/64 en
parallèle (textes longs enfin couverts, labels meilleurs qu'en 512).
Harness Go/Python : toujours 100 %.

## 2026-07-17 — Rétro finale v0.1.0
**Ce qui a marché** : le corpus oracle (tests Python → jsonl partagé) a porté
tout le TDD ; l'extraction AST/sed des patterns Python a évité les typos ; le
harness Go-vs-Python a immédiatement attrapé la seule vraie lacune de portage
(invalidation SSN) → 100 % d'accord final ; les 3 pièges RE2 prévus au plan
sont exactement ceux rencontrés.
**Ce qui a coûté** : versions onnxruntime ↔ binding (2 essais), modèle cased
lowercasé (zéro entité sans erreur), encodage cp1252 des .ps1.
**À faire ensuite** : locales UE (portage AST déjà outillé), lemmatisation
dans contextaware quand le NLP fournira les lemmes, PHONE_NUMBER via
nyaruka/phonenumbers, publier le module sur GitHub.

Noter ici ce qui a fonctionné ou non, AU MOMENT où ça mord. Dater chaque entrée.

## 2026-07-16 — bootstrap
- Go absent de la machine : installé via `winget install GoLang.Go` (go1.26.5).
  Penser à rafraîchir le PATH de la session PowerShell après winget.

## 2026-07-16 — core-types
- `go test -race` sous Windows exige cgo (gcc) : non installé → lancer les tests
  locaux sans -race, la CI ubuntu s'en charge.
- Offsets runes : `FindAllStringIndex` retourne des bytes ; conversion au point de
  match avec `utf8.RuneCountInString` — testé avec « Prénom/José » (multi-bytes).

## 2026-07-16 — generic-recognizers
- Comme prévu au plan : 3 regex Python non-RE2 (lookahead carte, lookbehind IP,
  backreference MAC). Parades : validation Go, netip, patterns scindés.
- Grosse regex (URL, ~6 Ko de TLD) : extraite par script depuis le source Python
  plutôt que recopiée — zéro risque de typo, régénérable.
- gopls dans l'IDE pointe sur le workspace presidigo (Python) : erreurs fantômes
  sur presidigo-go → ouvrir presidigo-go comme projet séparé ou ajouter un go.work.
