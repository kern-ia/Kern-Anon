//go:build onnx

package main

import (
	"context"
	"fmt"

	"github.com/YoLaub/PresidioGo/analyzer"
	"github.com/YoLaub/PresidioGo/nlp/onnx"
	"github.com/YoLaub/PresidioGo/recognizers/ner"
	"github.com/YoLaub/PresidioGo/registry"
)

func main() {
	reg := registry.Default("fr")
	reg.Add(ner.New("fr"))

	eng, err := analyzer.New(
		analyzer.WithRegistry(reg),
		analyzer.WithNlpEngine(onnx.New("models/bert-ner-multilingual")),
		analyzer.WithDefaultLanguage("fr"),
	)
	if err != nil {
		panic(err)
	}

	text := "Le client, Jean Dupont, domicilie a Vannes, a fourni son IBAN FR7630006000011234567890189 " +
		"et son email jean.dupont@example.com. Son epouse Marie Curie travaille chez Alpha Credit."

	results, err := eng.Analyze(context.Background(), text, analyzer.Language("fr"))
	if err != nil {
		panic(err)
	}
	fmt.Println("nb results:", len(results))
	for _, res := range results {
		extrait := string([]rune(text)[res.Start:res.End])
		fmt.Printf("%-13s [%4d:%4d] score=%.2f -> %q\n", res.EntityType, res.Start, res.End, res.Score, extrait)
	}
}
