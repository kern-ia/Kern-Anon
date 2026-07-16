# presidigo-go

Bibliothèque Go de détection et d'anonymisation de PII dans du texte.
Refonte Go idiomatique du cœur de [Presidio](https://github.com/data-privacy-stack/presidio)
(MIT) : recognizers extensibles (regex + checksum + NER), registry par langue/entité,
boost contextuel, opérateurs d'anonymisation.

```go
eng, _ := analyzer.New(analyzer.WithRegistry(registry.Default("en", "fr")))
results, _ := eng.Analyze(ctx, "Contact: jean@exemple.fr", analyzer.MinScore(0.4))

out, _ := anonymizer.New().Anonymize(text, results, map[string]anonymizer.Operator{
    "DEFAULT": anonymizer.Replace("<PII>"),
})
```

- **100 % Go pur par défaut** — le NER (BERT via ONNX) est opt-in : `go build -tags onnx`
- **Offsets en runes**, sûrs avec accents et emoji
- **Corpus oracle** issu des tests de Presidio pour garantir la non-régression

Statut : v0 en construction (voir `docs/PLAN.md` et `docs/index/`).

## Licence
MIT — dérivé de l'architecture de Microsoft Presidio (MIT).
