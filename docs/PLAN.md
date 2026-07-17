# Plan de refonte — presidigo-go
> Rédigé le 2026-07-16 — workflow greenfield-tdd-okf, Phase 1 (à valider avant toute ligne de code)

## 1. Contexte & objectif

Porter le cœur de Presidio (détection + anonymisation de PII dans du **texte**) en Go,
sous forme de **bibliothèque importable** (`go get github.com/YoLaub/PresidioGo`),
destinée à devenir un module in-process d'un projet Go existant.

Motivation validée : supprimer le sidecar Python (~500 Mo-1 Go RAM avec spaCy) et
l'aller-retour HTTP ; ~80 % de la valeur (recognizers pattern/checksum) est de la
logique pure portable ; performances regex supérieures en Go.

Le périmètre fonctionnel de référence est `presidigo/archi-output/INDEX.md`.

## 2. Décisions validées (2026-07-16)

| Décision | Choix |
|---|---|
| Périmètre v1 | Texte seul : analyzer + anonymizer (pas d'image/DICOM, pas de structured) |
| NER | ONNX embarqué (bert-base-NER converti ONNX, via onnxruntime/cgo) derrière une interface pluggable |
| Emplacement | Nouveau repo : `c:\Users\laube\WebstormProjects\presidigo-go` |
| Module Go | `github.com/YoLaub/PresidioGo` |
| API | Go idiomatique (PAS de compatibilité REST Presidio) |
| Locales v1 | Génériques → USA (portage) → France (création) → UE par vagues |

## 3. Références open source étudiées

| Projet | Constat |
|---|---|
| [aliengiraffe/deidentify](https://github.com/aliengiraffe/deidentify) | Go, anonymisation déterministe HMAC — bonne idée à reprendre pour l'opérateur hash, mais pas d'architecture extensible |
| [aavaz-ai/pii-scrubber](https://github.com/aavaz-ai/pii-scrubber) | Go, scrubbers regex prédéfinis — pas de registry, pas de scoring, pas de NER |
| Presidio Python (ce fork) | L'architecture de référence : registry + recognizers + context enhancement + operators |

Conclusion : aucune lib Go existante ne fournit registry extensible + scoring + NER
+ operators. La refonte est justifiée ; on s'inspire de deidentify pour le hash déterministe.

## 4. Architecture

```
presidigo-go/
├── pii/              # types partagés (aucune dépendance interne)
├── recognizer/       # interface Recognizer + PatternRecognizer générique
├── registry/         # Registry : add/remove/filtrage par langue & entité
├── recognizers/
│   ├── generic/      # creditcard(Luhn), email, iban(mod97), ip, url, crypto, phone…
│   ├── us/           # ssn, itin, passport, driver_license (portés du fork)
│   ├── fr/           # nir(sécu, clé 97), siren/siret(Luhn), cni, permis (à créer)
│   └── ...           # de/, it/, es/, pl/… par vagues
├── nlp/              # interface NlpEngine + NlpArtifacts ; noop inclus
│   └── onnx/         # BERT-NER via onnxruntime (cgo) — build tag `onnx`
├── contextaware/     # ContextEnhancer : boost de score par mots de contexte
├── analyzer/         # Engine : orchestre nlp → registry → contextaware → seuils
├── anonymizer/       # interface Operator + replace/mask/hash/redact/keep/encrypt
│                     # Engine.Anonymize + résolution des chevauchements
└── internal/testdata # corpus oracle extrait des tests Python du fork
```

### API publique visée

```go
// Analyse
eng, _ := analyzer.New(
    analyzer.WithRegistry(registry.Default("en", "fr")),
    analyzer.WithNlpEngine(onnx.New("models/bert-ner.onnx")), // ou nlp.NoOp{}
)
results, err := eng.Analyze(ctx, text,
    analyzer.Entities("EMAIL_ADDRESS", "FR_NIR"),
    analyzer.MinScore(0.4),
)

// Anonymisation
anon := anonymizer.New()
out, err := anon.Anonymize(text, results, map[string]anonymizer.Operator{
    "DEFAULT":       anonymizer.Replace("<PII>"),
    "EMAIL_ADDRESS": anonymizer.Mask('*', 4),
})
```

### Correspondance avec le fork Python

| Python (INDEX.md) | Go |
|---|---|
| AnalyzerEngine.analyze() | analyzer.Engine.Analyze() |
| RecognizerRegistry | registry.Registry |
| EntityRecognizer / PatternRecognizer | recognizer.Recognizer (interface) / recognizer.Pattern |
| NlpEngine (spaCy/Stanza/…) | nlp.NlpEngine (onnx / noop) |
| LemmaContextAwareEnhancer | contextaware.Enhancer |
| Operators + OperatorsFactory | anonymizer.Operator (interface, pas de factory — idiomatique) |
| ConflictResolutionStrategy | anonymizer : tri + fusion des spans qui se chevauchent |
| REGEX_TIMEOUT_SECONDS (anti-ReDoS) | inutile : le package regexp Go (RE2) est en temps linéaire garanti |

## 5. Modèle de données

```go
package pii

type Result struct {
    EntityType  string   // "EMAIL_ADDRESS", "FR_NIR", …
    Start, End  int      // offsets en bytes ET version runes exposée
    Score       float64  // 0.0–1.0
    Explanation *Explanation // recognizer, pattern, boost contextuel appliqué
}

type Pattern struct {
    Name  string
    Regex *regexp.Regexp
    Score float64
}
```

Points durs identifiés dès maintenant :
- **Offsets** : Python indexe en runes, Go en bytes → décision : API en runes
  (compatible corpus oracle), conversion interne unique.
- **Validation post-regex** : Luhn, mod-97 IBAN, clé NIR = fonctions `Validate(match) *bool`
  (nil = neutre, true = score→max, false = rejet), comme le fork.

## 6. Étapes (branches feature, TDD strict)

Chaque étape : tests écrits AVANT le code → suite verte → E2E → fiche OKF
`docs/index/<feature>.md` → merge `--no-ff` vers dev.

| # | Branche | Contenu | Dépend de |
|---|---|---|---|
| 1 | feature/bootstrap | go.mod, golangci-lint, CI GitHub Actions, CLAUDE.md, docs/index/ + retro.md, extraction du corpus oracle depuis les tests Python | — |
| 2 | feature/core-types | pii.Result, Pattern, interface Recognizer, PatternRecognizer, Registry | 1 |
| 3 | feature/generic-recognizers | les ~10 génériques avec validateurs (Luhn, mod-97…) | 2 |
| 4 | feature/anonymizer | Operators + Engine + chevauchements | 2 (parallélisable avec 3) |
| 5 | feature/context-enhancer | boost par mots de contexte + fenêtre | 3 |
| 6 | feature/nlp-onnx | interface NlpEngine, moteur ONNX (cgo, build tag), noop fallback | 2 (parallélisable avec 3-5) |
| 7 | feature/recognizers-us | portage SSN, ITIN, passeport, permis | 3 |
| 8 | feature/recognizers-fr | création NIR, SIREN/SIRET, CNI, permis, plaque | 3 |
| 9 | feature/recognizers-eu | DE, IT, ES, PL… par vagues selon besoin réel | 3 |
| 10 | release v0.1.0 | E2E oracle complet, rétro finale, merge dev→main, tag | tout |

Parallélisable : (3, 4, 6) après le socle 2 ; (7, 8, 9) après 3.

## 7. Vérification E2E

1. **Corpus oracle** : phrases synthétiques + cas des tests Python du fork
   (`internal/testdata/oracle.jsonl` : texte → entités attendues avec offsets).
2. **Harness de comparaison** : script qui lance le presidio-analyzer Python
   (docker compose du fork, port 5002) et compare ses sorties `/analyze` aux
   sorties Go sur le même corpus. Critère v0.1.0 : ≥ 95 % d'accord sur les
   entités pattern ; NER comparé à titre indicatif (moteurs différents).
3. **E2E réel** : petit programme Go d'exemple (`examples/`) importé comme le
   ferait ton projet — compile, détecte, anonymise, round-trip encrypt/decrypt.

## 8. Risques & parades

| Risque | Parade |
|---|---|
| cgo (onnxruntime) complique le build | build tag `onnx` : sans le tag, lib 100 % Go pur avec nlp.NoOp ; le NER devient opt-in |
| Divergence d'offsets runes/bytes | corpus oracle avec accents/emoji dès l'étape 2 |
| RE2 Go ne supporte pas lookahead/backreferences présents dans certaines regex Python | réécrire ces regex (validation en code Go) — recensées au portage, notées dans retro.md |
| Scope creep (66 locales) | vagues à la demande, chaque vague = une fiche OKF |

## 9. Hors périmètre v1 (explicite)

Image/DICOM, structured (CSV/JSON), serveur HTTP, batch engines, opérateur
AHDS Azure, multi-moteurs NLP. Extensible plus tard sans casser l'API (interfaces).
