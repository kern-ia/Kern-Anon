// Exemple : cycle complet détection → anonymisation.
package main

import (
	"context"
	"fmt"

	"github.com/YoLaub/presidigo-go/anonymizer"
	"github.com/YoLaub/presidigo-go/pii"
	"github.com/YoLaub/presidigo-go/recognizers/generic"
	"github.com/YoLaub/presidigo-go/registry"
)

func main() {
	reg := registry.New()
	for _, rec := range generic.All("fr") {
		reg.Add(rec)
	}

	text := "Prénom : José — email : info@presidio.site, carte 4012-8888-8888-1881, " +
		"IBAN GB29NWBK60161331926819, IP 192.168.0.1, mais pas 4012-8888-8888-1882."

	var results []pii.Result
	for _, rec := range reg.Get("fr") {
		found, err := rec.Analyze(context.Background(), text, nil)
		if err != nil {
			panic(err)
		}
		results = append(results, found...)
	}
	for _, res := range results {
		extrait := string([]rune(text)[res.Start:res.End])
		fmt.Printf("%-13s [%3d:%3d] score=%.2f → %q\n",
			res.EntityType, res.Start, res.End, res.Score, extrait)
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
