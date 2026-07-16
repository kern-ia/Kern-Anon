// Exemple minimal : détection d'emails via PatternRecognizer + Registry.
package main

import (
	"context"
	"fmt"
	"regexp"

	"github.com/YoLaub/presidigo-go/pii"
	"github.com/YoLaub/presidigo-go/recognizer"
	"github.com/YoLaub/presidigo-go/registry"
)

func main() {
	email, err := recognizer.NewPattern("EmailRecognizer", "EMAIL_ADDRESS", "fr",
		[]pii.Pattern{{
			Name:  "email-basic",
			Regex: regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`),
			Score: 0.6,
		}})
	if err != nil {
		panic(err)
	}

	reg := registry.New()
	reg.Add(email)

	text := "Prénom : José — email : info@presidio.site"
	for _, rec := range reg.Get("fr") {
		results, err := rec.Analyze(context.Background(), text, nil)
		if err != nil {
			panic(err)
		}
		for _, res := range results {
			extrait := string([]rune(text)[res.Start:res.End])
			fmt.Printf("%s [%d:%d] score=%.2f → %q\n",
				res.EntityType, res.Start, res.End, res.Score, extrait)
		}
	}
}
