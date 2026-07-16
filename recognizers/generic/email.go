package generic

import (
	"regexp"

	"github.com/YoLaub/presidigo-go/pii"
	"github.com/YoLaub/presidigo-go/recognizer"
)

// Regex portée telle quelle du fork (RE2-compatible). Divergence assumée :
// \w est ASCII en Go alors qu'il est Unicode en Python — les emails IDN
// bruts ne sont pas détectés (les formes punycode xn-- le sont).
var emailPattern = pii.Pattern{
	Name:  "Email (Medium)",
	Regex: regexp.MustCompile(`\b((([!#$%&'*+\-/=?^_` + "`" + `{|}~\w])|([!#$%&'*+\-/=?^_` + "`" + `{|}~\w][!#$%&'*+\-/=?^_` + "`" + `{|}~\.\w]{0,}[!#$%&'*+\-/=?^_` + "`" + `{|}~\w]))[@]\w+(?:-+\w+)*(?:\.\w+(?:-+\w+)*)+)\b`),
	Score: 0.5,
}

// NewEmail détecte les adresses email (EMAIL_ADDRESS).
func NewEmail(language string) *recognizer.PatternRecognizer {
	return mustPattern("EmailRecognizer", "EMAIL_ADDRESS", language,
		[]pii.Pattern{emailPattern})
}
