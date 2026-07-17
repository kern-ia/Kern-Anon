// Exemple : le pipeline complet via l'API publique — analyzer → anonymizer.
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

	text := "Prénom : José — email : info@presidio.site, carte 4012-8888-8888-1881, " +
		"IBAN GB29NWBK60161331926819, IP 192.168.0.1, mais pas 4012-8888-8888-1882."

	results, err := eng.Analyze(context.Background(), text, analyzer.MinScore(0.4))
	if err != nil {
		panic(err)
	}
	for _, res := range results {
		extrait := string([]rune(text)[res.Start:res.End])
		boost := ""
		if res.Explanation != nil && res.Explanation.ContextBoost > 0 {
			boost = fmt.Sprintf(" (boost contexte +%.2f)", res.Explanation.ContextBoost)
		}
		fmt.Printf("%-13s [%3d:%3d] score=%.2f%s → %q\n",
			res.EntityType, res.Start, res.End, res.Score, boost, extrait)
	}

	out, err := anonymizer.New().Anonymize(text, results, map[string]anonymizer.Operator{
		"CREDIT_CARD": anonymizer.Mask('*', 15, false),
		"DEFAULT":     anonymizer.Replace("<PII>"),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("\nAnonymisé :", out.Text)
}
