package generic

import (
	"regexp"

	"github.com/YoLaub/PresidioGo/pii"
	"github.com/YoLaub/PresidioGo/recognizer"
)

// Le pattern Python utilise une backreference \1 pour interdire les
// séparateurs mélangés (non supporté par RE2) : scindé en deux patterns,
// un par séparateur, qui préservent la même sémantique.
var macPatterns = []pii.Pattern{
	{
		Name:  "MAC_COLON",
		Regex: regexp.MustCompile(`\b[0-9A-Fa-f]{2}(?::[0-9A-Fa-f]{2}){5}\b`),
		Score: 0.6,
	},
	{
		Name:  "MAC_HYPHEN",
		Regex: regexp.MustCompile(`\b[0-9A-Fa-f]{2}(?:-[0-9A-Fa-f]{2}){5}\b`),
		Score: 0.6,
	},
	{
		Name:  "MAC_CISCO_DOT",
		Regex: regexp.MustCompile(`\b[0-9A-Fa-f]{4}\.[0-9A-Fa-f]{4}\.[0-9A-Fa-f]{4}\b`),
		Score: 0.6,
	},
}

// NewMAC détecte les adresses MAC (MAC_ADDRESS).
func NewMAC(language string) *recognizer.PatternRecognizer {
	return mustPattern("MacRecognizer", "MAC_ADDRESS", language, macPatterns,
		recognizer.WithContextWords("mac", "mac address", "hardware address", "physical address", "ethernet"))
}
