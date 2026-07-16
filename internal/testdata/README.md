# Corpus oracle

`oracle.jsonl` : une ligne JSON par cas — `text` + `expected` (entités attendues,
offsets **en runes**, `min_score`). Les cas proviennent des tests Python du fork
presidigo (`../presidigo/presidio-analyzer/tests/`) et de cas propres à ce repo
(offsets runes avec accents/emoji).

Règles :
- Données 100 % synthétiques (numéros de test normatifs : Luhn/mod-97 valides mais fictifs).
- Chaque feature de recognizer AJOUTE ses cas ici (positifs ET négatifs) avant d'écrire le code.
- Ne jamais supprimer un cas existant sans noter pourquoi dans docs/index/retro.md.
