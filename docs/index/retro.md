# Rétro continue — presidigo-go

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
