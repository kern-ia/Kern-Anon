//go:build onnx

// Exemple NER : pipeline complet avec le moteur ONNX (build tag onnx).
//
//	./scripts/download-model.ps1   (ou .sh)
//	ONNXRUNTIME_LIB=models/onnxruntime/onnxruntime.dll go run -tags onnx ./examples/ner
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/YoLaub/PresidioGo/analyzer"
	"github.com/YoLaub/PresidioGo/nlp/onnx"
	"github.com/YoLaub/PresidioGo/recognizers/ner"
	"github.com/YoLaub/PresidioGo/registry"
)

func main() {
	reg := registry.Default("en")
	reg.Add(ner.New("en"))

	eng, err := analyzer.New(
		analyzer.WithRegistry(reg),
		analyzer.WithNlpEngine(onnx.New("models/bert-ner")),
	)
	if err != nil {
		panic(err)
	}

	text := "My name is John Smith, I work at Microsoft in Seattle. " +
		"My card is 4012-8888-8888-1881 and my email is john@example.org."

	show(eng, text)

	// Texte long (> 512 tokens) : les entités en fin de document ne sont
	// visibles que grâce au fenêtrage chevauchant.
	long := strings.Repeat("The quarterly report covers revenue, costs and forecasts for the next period. ", 60) +
		"The final section was written by Marie Curie at the University of Warsaw."
	fmt.Println("\n--- texte long (fenêtrage) ---")
	show(eng, long)
}

func show(eng *analyzer.Engine, text string) {
	results, err := eng.Analyze(context.Background(), text, analyzer.MinScore(0.4))
	if err != nil {
		panic(err)
	}
	for _, res := range results {
		extrait := string([]rune(text)[res.Start:res.End])
		fmt.Printf("%-13s [%4d:%4d] score=%.2f → %q\n",
			res.EntityType, res.Start, res.End, res.Score, extrait)
	}
}
