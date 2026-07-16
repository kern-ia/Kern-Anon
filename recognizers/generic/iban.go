package generic

import (
	"regexp"

	"github.com/YoLaub/presidigo-go/pii"
	"github.com/YoLaub/presidigo-go/recognizer"
)

// Le fork Python porte un pattern par pays (iban_patterns.py) ; v1 Go :
// pattern générique + validation mod 97, qui porte l'essentiel de la
// fiabilité. Les longueurs par pays pourront s'ajouter ensuite.
var ibanPattern = pii.Pattern{
	Name:  "IBAN (generic)",
	Regex: regexp.MustCompile(`\b[A-Z]{2}\d{2}[A-Z0-9]{11,30}\b`),
	Score: 0.5,
}

// NewIban détecte les IBAN (IBAN_CODE), validés par ISO 7064 mod 97-10.
func NewIban(language string) *recognizer.PatternRecognizer {
	return mustPattern("IbanRecognizer", "IBAN_CODE", language,
		[]pii.Pattern{ibanPattern},
		recognizer.WithValidate(func(match string) *bool {
			ok := IbanMod97(match)
			return &ok
		}))
}
