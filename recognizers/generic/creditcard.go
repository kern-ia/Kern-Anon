package generic

import (
	"regexp"
	"strings"

	"github.com/YoLaub/presidigo-go/pii"
	"github.com/YoLaub/presidigo-go/recognizer"
)

// La regex Python commence par le lookahead négatif (?!1\d{12}(?!\d))
// (exclut 13 chiffres commençant par 1), non supporté par RE2 :
// l'exclusion est portée dans la validation ci-dessous.
var creditCardPattern = pii.Pattern{
	Name:  "All Credit Cards (weak)",
	Regex: regexp.MustCompile(`\b((4\d{3})|(5[0-5]\d{2})|(6\d{3})|(1\d{3})|(3\d{3}))[- ]?(\d{3,4})[- ]?(\d{3,4})[- ]?(\d{3,5})\b`),
	Score: 0.3,
}

// NewCreditCard détecte les numéros de carte bancaire (CREDIT_CARD),
// validés par somme de contrôle Luhn.
func NewCreditCard(language string) *recognizer.PatternRecognizer {
	return mustPattern("CreditCardRecognizer", "CREDIT_CARD", language,
		[]pii.Pattern{creditCardPattern},
		recognizer.WithValidate(func(match string) *bool {
			sanitized := strings.NewReplacer("-", "", " ", "").Replace(match)
			ok := !(len(sanitized) == 13 && sanitized[0] == '1') && Luhn(sanitized)
			return &ok
		}))
}
