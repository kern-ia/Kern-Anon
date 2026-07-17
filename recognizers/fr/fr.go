// Package fr fournit les recognizers PII français — créés pour presidigo-go,
// le fork Presidio n'ayant pas de recognizers France :
// NIR (sécurité sociale, clé 97), SIREN/SIRET (Luhn), plaque SIV, téléphone.
package fr

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/YoLaub/presidigo-go/pii"
	"github.com/YoLaub/presidigo-go/recognizer"
	"github.com/YoLaub/presidigo-go/recognizers/generic"
)

// All retourne les recognizers français (langue "fr").
func All() []recognizer.Recognizer {
	return []recognizer.Recognizer{
		NewNIR(),
		NewSIREN(),
		NewSIRET(),
		NewLicensePlate(),
		NewPhone(),
	}
}

func must(name, entity string, patterns []pii.Pattern, opts ...recognizer.Option) *recognizer.PatternRecognizer {
	r, err := recognizer.NewPattern(name, entity, "fr", patterns, opts...)
	if err != nil {
		panic(err)
	}
	return r
}

func stripSpaces(s string) string {
	return strings.NewReplacer(" ", "", ".", "", "-", "").Replace(s)
}

// NewNIR détecte le numéro de sécurité sociale français (FR_NIR),
// validé par la clé de contrôle (97 - n mod 97, Corse 2A→19 / 2B→18).
func NewNIR() *recognizer.PatternRecognizer {
	return must("FrNirRecognizer", "FR_NIR", []pii.Pattern{{
		Name:  "NIR (weak)",
		Regex: regexp.MustCompile(`\b[12][ ]?\d{2}[ ]?\d{2}[ ]?(\d{2}|2[AB])[ ]?\d{3}[ ]?\d{3}[ ]?\d{2}\b`),
		Score: 0.3,
	}},
		recognizer.WithContextWords("sécurité sociale", "sécu", "nir", "insee", "assuré"),
		recognizer.WithValidate(func(match string) *bool {
			ok := NirKey(stripSpaces(match))
			return &ok
		}))
}

// NirKey vérifie la clé d'un NIR de 15 caractères (sans espaces) :
// clé = 97 - (13 premiers caractères mod 97), avec 2A→19 et 2B→18.
func NirKey(nir string) bool {
	if len(nir) != 15 {
		return false
	}
	body := strings.ToUpper(nir[:13])
	body = strings.Replace(body, "2A", "19", 1)
	body = strings.Replace(body, "2B", "18", 1)
	n, err := strconv.ParseInt(body, 10, 64)
	if err != nil {
		return false
	}
	key, err := strconv.Atoi(nir[13:])
	if err != nil {
		return false
	}
	return key == int(97-n%97)
}

// NewSIREN détecte les numéros SIREN (FR_SIREN), validés par Luhn.
// Limite connue : les SIREN de La Poste (356000000) ne passent pas Luhn.
func NewSIREN() *recognizer.PatternRecognizer {
	return must("FrSirenRecognizer", "FR_SIREN", []pii.Pattern{{
		Name:  "SIREN (weak)",
		Regex: regexp.MustCompile(`\b\d{3}[ ]?\d{3}[ ]?\d{3}\b`),
		Score: 0.3,
	}},
		recognizer.WithContextWords("siren", "entreprise", "société", "rcs"),
		recognizer.WithValidate(func(match string) *bool {
			ok := generic.Luhn(stripSpaces(match))
			return &ok
		}))
}

// NewSIRET détecte les numéros SIRET (FR_SIRET), validés par Luhn sur 14 chiffres.
func NewSIRET() *recognizer.PatternRecognizer {
	return must("FrSiretRecognizer", "FR_SIRET", []pii.Pattern{{
		Name:  "SIRET (weak)",
		Regex: regexp.MustCompile(`\b\d{3}[ ]?\d{3}[ ]?\d{3}[ ]?\d{5}\b`),
		Score: 0.3,
	}},
		recognizer.WithContextWords("siret", "établissement", "facture"),
		recognizer.WithValidate(func(match string) *bool {
			ok := generic.Luhn(stripSpaces(match))
			return &ok
		}))
}

// NewLicensePlate détecte les plaques d'immatriculation SIV (FR_LICENSE_PLATE),
// format AA-123-AA, lettres I, O et U exclues.
func NewLicensePlate() *recognizer.PatternRecognizer {
	return must("FrLicensePlateRecognizer", "FR_LICENSE_PLATE", []pii.Pattern{{
		Name:  "Plaque SIV",
		Regex: regexp.MustCompile(`\b[A-HJ-NP-TV-Z]{2}-\d{3}-[A-HJ-NP-TV-Z]{2}\b`),
		Score: 0.4,
	}},
		recognizer.WithContextWords("immatriculation", "plaque", "véhicule", "voiture"))
}

// NewPhone détecte les numéros de téléphone français (FR_PHONE_NUMBER),
// formats nationaux (06 01 02 03 04) et internationaux (+33 6 …).
func NewPhone() *recognizer.PatternRecognizer {
	return must("FrPhoneRecognizer", "FR_PHONE_NUMBER", []pii.Pattern{
		{
			Name:  "Téléphone national",
			Regex: regexp.MustCompile(`\b0[1-9](?:[ .-]?\d{2}){4}\b`),
			Score: 0.4,
		},
		{
			Name:  "Téléphone international",
			Regex: regexp.MustCompile(`(?:\+33|0033)[ .-]?[1-9](?:[ .-]?\d{2}){4}\b`),
			Score: 0.5,
		},
	},
		recognizer.WithContextWords("téléphone", "tél", "portable", "mobile", "appelle", "joindre"))
}
