package generic

import (
	"net/netip"
	"regexp"

	"github.com/YoLaub/presidigo-go/pii"
	"github.com/YoLaub/presidigo-go/recognizer"
)

// Les patterns IPv6 du fork reposent sur des lookbehind/lookahead (non-RE2).
// Stratégie Go : regex candidates larges + validation par net/netip, plus
// fiable qu'une regex exhaustive. La validation est neutre (nil) quand
// l'adresse parse — le score du pattern est conservé, comme dans le fork
// qui ne valide pas les IP.
var ipPatterns = []pii.Pattern{
	{
		Name:  "IPv4",
		Regex: regexp.MustCompile(`\b(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`),
		Score: 0.6,
	},
	{
		Name:  "IPv6 (candidate)",
		Regex: regexp.MustCompile(`(?:[0-9A-Fa-f]{1,4}:){2,7}[0-9A-Fa-f]{0,4}(?::[0-9A-Fa-f]{1,4})*`),
		Score: 0.6,
	},
}

// NewIP détecte les adresses IPv4/IPv6 (IP_ADDRESS), validées par netip.
func NewIP(language string) *recognizer.PatternRecognizer {
	return mustPattern("IpRecognizer", "IP_ADDRESS", language, ipPatterns,
		recognizer.WithContextWords("ip", "ipv4", "ipv6"),
		recognizer.WithValidate(func(match string) *bool {
			if _, err := netip.ParseAddr(match); err != nil {
				invalid := false
				return &invalid
			}
			return nil // valide : score du pattern conservé
		}))
}
