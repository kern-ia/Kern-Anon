// Exemple : démonstration du bon fonctionnement de l'anonymizer sur des
// textes réalistes (fr + en), avec plusieurs opérateurs (Mask, Hash, Redact,
// Replace, Encrypt/Decrypt) et un round-trip de déanonymisation.
package main

import (
	"context"
	"fmt"

	"github.com/YoLaub/PresidioGo/analyzer"
	"github.com/YoLaub/PresidioGo/anonymizer"
	"github.com/YoLaub/PresidioGo/registry"
)

type sample struct {
	language string
	text     string
}

func main() {
	samples := []sample{
		{"fr", "Contactez Léa Dupont : lea.dupont@exemple.fr ou au 06 01 02 03 04. " +
			"Virement depuis IBAN GB29NWBK60161331926819, carte 4012-8888-8888-1881, " +
			"SIREN 552100554. Numéro de sécu : 1 85 05 78 006 084 91."},
		{"en", "Reach me at john.smith@example.com or my SSN is 216-09-1234. " +
			"Server IP 192.168.0.1, visit https://example.com for details."},
	}

	key := []byte("0123456789abcdef0123456789abcdef") // 32 octets → AES-256-GCM
	enc, err := anonymizer.Encrypt(key)
	if err != nil {
		panic(err)
	}
	dec, err := anonymizer.Decrypt(key)
	if err != nil {
		panic(err)
	}

	ops := map[string]anonymizer.Operator{
		"CREDIT_CARD":   anonymizer.Mask('*', 4, true), // garde les 4 derniers chiffres
		"US_SSN":        anonymizer.Hash(),
		"FR_NIR":        anonymizer.Hash(),
		"IBAN_CODE":     anonymizer.Redact(),
		"EMAIL_ADDRESS": enc, // réversible, voir round-trip plus bas
		"DEFAULT":       anonymizer.Replace("<PII>"),
	}
	deanonOps := map[string]anonymizer.Operator{
		"EMAIL_ADDRESS": dec,
		"DEFAULT":       anonymizer.Keep(),
	}

	anon := anonymizer.New()

	for _, s := range samples {
		eng, err := analyzer.New(
			analyzer.WithRegistry(registry.Default(s.language)),
			analyzer.WithDefaultLanguage(s.language),
		)
		if err != nil {
			panic(err)
		}

		results, err := eng.Analyze(context.Background(), s.text, analyzer.MinScore(0.4))
		if err != nil {
			panic(err)
		}

		fmt.Printf("=== [%s] texte original ===\n%s\n\n", s.language, s.text)

		fmt.Println("Entités détectées :")
		for _, r := range results {
			extrait := string([]rune(s.text)[r.Start:r.End])
			fmt.Printf("  %-14s [%3d:%3d] score=%.2f → %q\n", r.EntityType, r.Start, r.End, r.Score, extrait)
		}

		out, err := anon.Anonymize(s.text, results, ops)
		if err != nil {
			panic(err)
		}
		fmt.Printf("\nAnonymisé :\n%s\n\n", out.Text)

		back, err := anon.Deanonymize(out.Text, out.Items, deanonOps)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Round-trip (email déchiffré) :\n%s\n", back.Text)
		fmt.Println("-------------------------------------------------------------")
	}
}
