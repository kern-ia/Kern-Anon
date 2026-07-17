# PresidioGo

Bibliothèque Go de **détection** (analyzer) et d'**anonymisation** (anonymizer) de
données personnelles (PII) dans du texte. Refonte Go idiomatique du cœur de
[Presidio](https://github.com/data-privacy-stack/presidio) (MIT), pensée pour être
importée comme module dans un projet Go — pas comme un microservice à déployer.

```go
eng, _ := analyzer.New(analyzer.WithRegistry(registry.Default("fr")))
results, _ := eng.Analyze(ctx, "Contact : jean@exemple.fr", analyzer.MinScore(0.4))

out, _ := anonymizer.New().Anonymize(text, results, map[string]anonymizer.Operator{
    "DEFAULT": anonymizer.Replace("<PII>"),
})
```

## Ce que fait PresidioGo

Le pipeline complet tient en deux moteurs :

1. **Analyzer** — parcourt le texte avec des *recognizers* (regex + validation par
   checksum, et NER optionnel) enregistrés par langue dans un *registry*. Chaque
   résultat porte un type d'entité, une position `[Start:End]` **en runes** (sûre
   avec accents et emoji), un score et une explication. Un *boost contextuel*
   augmente le score quand un mot de contexte (« iban », « carte », « ssn »…)
   précède l'entité ; un seuil `MinScore` filtre le bruit.
2. **Anonymizer** — remplace les spans détectés selon un opérateur choisi par
   entité : `Replace`, `Redact`, `Mask`, `Hash`, `Keep`, `Custom`, ou
   `Encrypt`/`Decrypt` (AES-GCM, avec `Deanonymize` pour le round-trip). Les
   chevauchements sont résolus avant substitution : l'entité au meilleur score gagne.

Un match regex peut être **validé** (checksum OK → score maximal), **invalidé**
(checksum KO → rejeté) ou laissé au score du pattern — exactement la sémantique
de Presidio.

### Entités supportées

| Groupe | Entités |
|---|---|
| Génériques | `CREDIT_CARD` (Luhn), `EMAIL_ADDRESS`, `IBAN_CODE` (mod-97), `IP_ADDRESS` (v4/v6), `URL`, `MAC_ADDRESS`, `CRYPTO` (base58check, bech32) |
| États-Unis | `US_SSN`, `US_ITIN`, `US_PASSPORT`, `US_DRIVER_LICENSE`, `US_BANK_NUMBER`, `ABA_ROUTING_NUMBER`, `US_NPI` (Luhn), `US_MBI`, `MEDICAL_LICENSE` (DEA) |
| France | `FR_NIR` (clé 97, Corse incluse), `FR_SIREN`, `FR_SIRET` (Luhn), `FR_LICENSE_PLATE` (SIV), `FR_PHONE_NUMBER` |
| NER (opt-in) | `PERSON`, `LOCATION`, `ORGANIZATION`, `NRP` via BERT-NER (ONNX) |

`registry.Default("en", "fr")` assemble génériques + recognizers de chaque langue
demandée. Ajouter un recognizer maison = implémenter une petite interface et
l'enregistrer.

## Pourquoi un portage Go ?

Presidio est une excellente référence, mais c'est un **service Python** : pour
l'utiliser depuis un backend Go il faut déployer un conteneur, gérer un runtime
Python + spaCy, et payer un aller-retour HTTP par analyse. PresidioGo inverse
le modèle :

- **Une bibliothèque, pas un service.** `go get`, un import, zéro
  infrastructure. L'analyse se fait dans le processus, sans latence réseau ni
  sérialisation.
- **100 % Go pur par défaut.** Sans le tag `onnx`, aucune dépendance cgo ni
  binaire externe : la lib se cross-compile et s'embarque dans un binaire
  statique unique (CLI, Lambda, sidecar…).
- **Rapide et sobre.** Regex RE2 en temps linéaire garanti, recognizers exécutés
  en parallèle (goroutines), fenêtre contextuelle bornée. Mesuré sur un document
  de 30 Ko avec les 28 recognizers en+fr : **13,3 ms et 3,8 Mo alloués par
  analyse** (contre 550 ms / 534 Mo avant optimisation — ×41 en vitesse, ÷139
  en mémoire).
- **Fidèle à l'original.** Les patterns et règles de validation sont extraits du
  code Python (par AST, pas retranscrits à la main) et un corpus oracle
  (`internal/testdata/oracle.jsonl`) issu des tests de Presidio verrouille la
  non-régression. Le harness E2E compare span par span avec le
  presidio-analyzer Python : **100 % d'accord** sur les entités communes.
- **NER maîtrisé.** Là où Presidio embarque spaCy, PresidioGo propose un NER
  opt-in : tokenizer WordPiece et agrégation BIO en pur Go, inférence
  BERT-NER via ONNX Runtime uniquement si vous activez le tag de build.

## Installation

Go ≥ 1.26.

```sh
go get github.com/YoLaub/PresidioGo
```

C'est tout pour l'usage par défaut (regex + checksums + contexte, pur Go).

### Démarrage rapide

```go
package main

import (
    "context"
    "fmt"

    "github.com/YoLaub/PresidioGo/analyzer"
    "github.com/YoLaub/PresidioGo/anonymizer"
    "github.com/YoLaub/PresidioGo/registry"
)

func main() {
    eng, err := analyzer.New(
        analyzer.WithRegistry(registry.Default("fr")),
        analyzer.WithDefaultLanguage("fr"),
    )
    if err != nil {
        panic(err)
    }

    text := "Email : info@presidio.site, carte 4012-8888-8888-1881, IBAN GB29NWBK60161331926819"

    results, _ := eng.Analyze(context.Background(), text, analyzer.MinScore(0.4))

    out, _ := anonymizer.New().Anonymize(text, results, map[string]anonymizer.Operator{
        "CREDIT_CARD": anonymizer.Mask('*', 15, false),
        "DEFAULT":     anonymizer.Replace("<PII>"),
    })
    fmt.Println(out.Text)
}
```

Exemple complet exécutable : `go run ./examples/basic`.

### NER (optionnel, tag `onnx`)

Le NER ajoute la détection de personnes, lieux et organisations via
BERT-NER (modèle quantisé int8, ~110 Mo). Il nécessite cgo et ONNX Runtime :

```sh
# 1. Télécharger le modèle + onnxruntime 1.26.x
./scripts/download-model.sh        # ou scripts/download-model.ps1 sous Windows

# 2. Indiquer la bibliothèque ONNX Runtime
export ONNXRUNTIME_LIB=/chemin/vers/libonnxruntime.so

# 3. Compiler avec le tag
go build -tags onnx ./...
go run -tags onnx ./examples/ner
```

Sans le tag, ces packages sont simplement absents du build : rien à installer,
rien à configurer.

## Développement

```sh
go test ./...          # tests (TDD, corpus oracle)
golangci-lint run      # lint
go run ./internal/oracleharness   # E2E vs presidio-analyzer Python (Docker, port 5002)
```

Statut : v0.2 — pipeline complet (analyzer, anonymizer, 21 recognizers en/fr,
NER ONNX, harness oracle). Feuille de route : `docs/PLAN.md`, fiches par
feature : `docs/index/`.

## Licence

MIT — dérivé de l'architecture de Microsoft Presidio (MIT).
