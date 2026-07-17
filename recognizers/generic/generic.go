// Package generic fournit les recognizers PII indépendants du pays,
// portés des predefined_recognizers/generic du fork Presidio :
// carte bancaire, email, IBAN, IP, URL, MAC, crypto.
//
// Les regex Python utilisant lookahead/lookbehind/backreferences (non
// supportés par RE2) sont réécrites : la contrainte est déplacée dans la
// fonction de validation Go (voir chaque fichier).
package generic

import (
	"github.com/YoLaub/PresidioGo/pii"
	"github.com/YoLaub/PresidioGo/recognizer"
)

// All retourne les recognizers génériques pour la langue donnée.
func All(language string) []recognizer.Recognizer {
	return []recognizer.Recognizer{
		NewCreditCard(language),
		NewEmail(language),
		NewIban(language),
		NewIP(language),
		NewURL(language),
		NewMAC(language),
		NewCrypto(language),
	}
}

// mustPattern construit un PatternRecognizer dont les patterns sont connus
// à la compilation : une erreur est un bug de programmation.
func mustPattern(name, entity, language string, patterns []pii.Pattern, opts ...recognizer.Option) *recognizer.PatternRecognizer {
	r, err := recognizer.NewPattern(name, entity, language, patterns, opts...)
	if err != nil {
		panic(err)
	}
	return r
}
