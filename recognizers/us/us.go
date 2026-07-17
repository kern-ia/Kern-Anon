// Package us fournit les recognizers PII des États-Unis, portés de
// predefined_recognizers/country_specific/us du fork Presidio :
// SSN, ITIN, passeport, permis de conduire, compte bancaire, routing ABA,
// NPI, MBI Medicare, licence médicale DEA.
package us

import (
	"regexp"
	"strings"

	"github.com/YoLaub/PresidioGo/pii"
	"github.com/YoLaub/PresidioGo/recognizer"
	"github.com/YoLaub/PresidioGo/recognizers/generic"
)

// All retourne les recognizers US pour la langue donnée.
func All(language string) []recognizer.Recognizer {
	return []recognizer.Recognizer{
		NewSSN(language),
		NewITIN(language),
		NewPassport(language),
		NewDriverLicense(language),
		NewBankNumber(language),
		NewAbaRouting(language),
		NewNPI(language),
		NewMBI(language),
		NewMedicalLicense(language),
	}
}

func must(name, entity, language string, patterns []pii.Pattern, opts ...recognizer.Option) *recognizer.PatternRecognizer {
	r, err := recognizer.NewPattern(name, entity, language, patterns, opts...)
	if err != nil {
		panic(err)
	}
	return r
}

func sanitize(s string) string {
	return strings.NewReplacer("-", "", " ", "").Replace(s)
}

// NewSSN détecte les numéros de sécurité sociale US (US_SSN).
func NewSSN(language string) *recognizer.PatternRecognizer {
	return must("UsSsnRecognizer", "US_SSN", language, []pii.Pattern{
		{Name: "SSN1 (very weak)", Regex: regexp.MustCompile(`\b([0-9]{5})-([0-9]{4})\b`), Score: 0.05},
		{Name: "SSN2 (very weak)", Regex: regexp.MustCompile(`\b([0-9]{3})-([0-9]{6})\b`), Score: 0.05},
		{Name: "SSN3 (very weak)", Regex: regexp.MustCompile(`\b(([0-9]{3})-([0-9]{2})-([0-9]{4}))\b`), Score: 0.05},
		{Name: "SSN4 (very weak)", Regex: regexp.MustCompile(`\b[0-9]{9}\b`), Score: 0.05},
		{Name: "SSN5 (medium)", Regex: regexp.MustCompile(`\b([0-9]{3})[- .]([0-9]{2})[- .]([0-9]{4})\b`), Score: 0.5},
	},
		recognizer.WithContextWords("social", "security", "ssn", "ssns", "ssid"),
		recognizer.WithValidate(ssnValidate))
}

// ssnValidate porte l'invalidate_result du fork : rejet des délimiteurs
// mélangés, chiffres tous identiques, groupes à zéros, zones 000/666 et
// SSN canoniques publiés. Sinon neutre (score du pattern conservé).
func ssnValidate(match string) *bool {
	reject := false
	delims := map[rune]bool{}
	var digits []byte
	for _, c := range match {
		switch {
		case c == '.' || c == '-' || c == ' ':
			delims[c] = true
		case c >= '0' && c <= '9':
			digits = append(digits, byte(c))
		}
	}
	d := string(digits)
	switch {
	case len(delims) > 1: // délimiteurs mélangés
		reject = true
	case len(d) > 0 && strings.Count(d, d[:1]) == len(d): // tous identiques
		reject = true
	case len(d) == 9 && (d[3:5] == "00" || d[5:] == "0000"): // groupes à zéros
		reject = true
	case len(d) >= 3 && (d[:3] == "000" || d[:3] == "666"): // zone jamais émise
		reject = true
	case d == "123456789" || d == "987654320" || d == "078051120": // exemples publiés
		reject = true
	}
	if reject {
		invalid := false
		return &invalid
	}
	return nil
}

// NewITIN détecte les Individual Taxpayer Identification Numbers (US_ITIN).
func NewITIN(language string) *recognizer.PatternRecognizer {
	return must("UsItinRecognizer", "US_ITIN", language, []pii.Pattern{
		{Name: "Itin (very weak)", Regex: regexp.MustCompile(`\b9\d{2}[- ](5\d|6[0-5]|7\d|8[0-8]|9([0-2]|[4-9]))\d{4}\b|\b9\d{2}(5\d|6[0-5]|7\d|8[0-8]|9([0-2]|[4-9]))[- ]\d{4}\b`), Score: 0.05},
		{Name: "Itin (weak)", Regex: regexp.MustCompile(`\b9\d{2}(5\d|6[0-5]|7\d|8[0-8]|9([0-2]|[4-9]))\d{4}\b`), Score: 0.3},
		{Name: "Itin (medium)", Regex: regexp.MustCompile(`\b9\d{2}[- ](5\d|6[0-5]|7\d|8[0-8]|9([0-2]|[4-9]))[- ]\d{4}\b`), Score: 0.5},
	}, recognizer.WithContextWords("individual", "taxpayer", "itin", "tax", "payer", "taxid", "tin"))
}

// NewPassport détecte les numéros de passeport US (US_PASSPORT).
func NewPassport(language string) *recognizer.PatternRecognizer {
	return must("UsPassportRecognizer", "US_PASSPORT", language, []pii.Pattern{
		{Name: "Passport (very weak)", Regex: regexp.MustCompile(`(\b[0-9]{9}\b)`), Score: 0.05},
		{Name: "Passport Next Generation (very weak)", Regex: regexp.MustCompile(`(\b[A-Z][0-9]{8}\b)`), Score: 0.1},
	}, recognizer.WithContextWords("us", "united", "states", "passport", "passport#", "travel", "document"))
}

// NewDriverLicense détecte les permis de conduire US (US_DRIVER_LICENSE).
// La première alternative contient « A-Z] » sans crochet ouvrant : bug
// présent dans le fork, conservé tel quel pour l'iso-comportement.
func NewDriverLicense(language string) *recognizer.PatternRecognizer {
	return must("UsLicenseRecognizer", "US_DRIVER_LICENSE", language, []pii.Pattern{
		{Name: "Driver License - Alphanumeric (weak)", Regex: regexp.MustCompile(`\b([A-Z][0-9]{3,6}|[A-Z][0-9]{5,9}|[A-Z][0-9]{6,8}|[A-Z][0-9]{4,8}|[A-Z][0-9]{9,11}|[A-Z]{1,2}[0-9]{5,6}|H[0-9]{8}|V[0-9]{6}|X[0-9]{8}|A-Z]{2}[0-9]{2,5}|[A-Z]{2}[0-9]{3,7}|[0-9]{2}[A-Z]{3}[0-9]{5,6}|[A-Z][0-9]{13,14}|[A-Z][0-9]{18}|[A-Z][0-9]{6}R|[A-Z][0-9]{9}|[A-Z][0-9]{1,12}|[0-9]{9}[A-Z]|[A-Z]{2}[0-9]{6}[A-Z]|[0-9]{8}[A-Z]{2}|[0-9]{3}[A-Z]{2}[0-9]{4}|[A-Z][0-9][A-Z][0-9][A-Z]|[0-9]{7,8}[A-Z])\b`), Score: 0.3},
		{Name: "Driver License - Digits (very weak)", Regex: regexp.MustCompile(`\b([0-9]{6,14}|[0-9]{16})\b`), Score: 0.01},
	}, recognizer.WithContextWords("driver", "license", "permit", "lic", "identification", "dls", "cdls", "lic#", "driving"))
}

// NewBankNumber détecte les numéros de compte bancaire US (US_BANK_NUMBER).
func NewBankNumber(language string) *recognizer.PatternRecognizer {
	return must("UsBankRecognizer", "US_BANK_NUMBER", language, []pii.Pattern{
		{Name: "Bank Account (weak)", Regex: regexp.MustCompile(`\b[0-9]{8,17}\b`), Score: 0.05},
	}, recognizer.WithContextWords("check", "account", "account#", "acct", "bank", "save", "debit"))
}

// NewAbaRouting détecte les ABA routing numbers (ABA_ROUTING_NUMBER),
// validés par la somme pondérée 3-7-1.
func NewAbaRouting(language string) *recognizer.PatternRecognizer {
	return must("AbaRoutingRecognizer", "ABA_ROUTING_NUMBER", language, []pii.Pattern{
		{Name: "ABA routing number (weak)", Regex: regexp.MustCompile(`\b[0123678]\d{8}\b`), Score: 0.05},
		{Name: "ABA routing number", Regex: regexp.MustCompile(`\b[0123678]\d{3}-\d{4}-\d\b`), Score: 0.3},
	},
		recognizer.WithContextWords("aba", "routing", "abarouting", "association", "bankrouting"),
		recognizer.WithValidate(func(match string) *bool {
			ok := AbaChecksum(sanitize(match))
			return &ok
		}))
}

// AbaChecksum vérifie la somme pondérée (3,7,1) d'un routing number à 9 chiffres.
func AbaChecksum(value string) bool {
	if len(value) != 9 {
		return false
	}
	weights := [9]int{3, 7, 1, 3, 7, 1, 3, 7, 1}
	sum := 0
	for i := 0; i < 9; i++ {
		c := value[i]
		if c < '0' || c > '9' {
			return false
		}
		sum += int(c-'0') * weights[i]
	}
	return sum%10 == 0
}

// NewNPI détecte les National Provider Identifiers (US_NPI), validés par
// Luhn avec le préfixe CMS « 80840 ».
func NewNPI(language string) *recognizer.PatternRecognizer {
	return must("UsNpiRecognizer", "US_NPI", language, []pii.Pattern{
		{Name: "NPI (weak)", Regex: regexp.MustCompile(`\b[12]\d{9}\b`), Score: 0.1},
		{Name: "NPI (medium)", Regex: regexp.MustCompile(`\b[12]\d{3}[ -]\d{3}[ -]\d{3}\b`), Score: 0.4},
	},
		recognizer.WithContextWords("npi", "national provider", "provider", "npi number", "provider id", "provider identifier", "taxonomy"),
		recognizer.WithValidate(func(match string) *bool {
			ok := NpiChecksum(sanitize(match))
			return &ok
		}))
}

// NpiChecksum valide un NPI : rejet des corps dégénérés (chiffres tous
// identiques, invalidate du fork) puis Luhn sur « 80840 » + NPI.
func NpiChecksum(value string) bool {
	if len(value) < 2 {
		return false
	}
	body := value[:len(value)-1]
	if strings.Count(body, body[:1]) == len(body) {
		return false
	}
	return generic.Luhn("80840" + value)
}

// mbi : lettres valides CMS (sans S, L, O, I, B, Z).
const mbiAlpha = "ACDEFGHJKMNPQRTUVWXY"

// NewMBI détecte les Medicare Beneficiary Identifiers (US_MBI).
func NewMBI(language string) *recognizer.PatternRecognizer {
	num, alpha, alnum := `[0-9]`, `[`+mbiAlpha+`]`, `[0-9`+mbiAlpha+`]`
	noDash := num + alpha + alnum + num + alpha + alnum + num + alpha + alpha + num + num
	withDash := num + alpha + alnum + num + `-` + alpha + alnum + num + `-` + alpha + alpha + num + num
	return must("UsMbiRecognizer", "US_MBI", language, []pii.Pattern{
		{Name: "MBI (weak)", Regex: regexp.MustCompile(`\b` + noDash + `\b`), Score: 0.3},
		{Name: "MBI (medium)", Regex: regexp.MustCompile(`\b` + withDash + `\b`), Score: 0.5},
	}, recognizer.WithContextWords("medicare", "mbi", "beneficiary", "cms", "medicaid", "hic", "hicn"))
}

// NewMedicalLicense détecte les numéros de certificat DEA (MEDICAL_LICENSE),
// validés par la somme de contrôle DEA.
func NewMedicalLicense(language string) *recognizer.PatternRecognizer {
	return must("MedicalLicenseRecognizer", "MEDICAL_LICENSE", language, []pii.Pattern{
		{Name: "USA DEA Certificate Number (weak)", Regex: regexp.MustCompile(`[abcdefghjklmprstuxABCDEFGHJKLMPRSTUX]{1}[a-zA-Z]{1}\d{7}|[abcdefghjklmprstuxABCDEFGHJKLMPRSTUX]{1}9\d{7}`), Score: 0.4},
	},
		recognizer.WithContextWords("medical", "certificate", "DEA"),
		recognizer.WithValidate(func(match string) *bool {
			ok := DeaChecksum(sanitize(match))
			return &ok
		}))
}

// DeaChecksum vérifie la somme de contrôle DEA : sur les 7 chiffres qui
// suivent les 2 lettres, (d1+d3+d5) + 2*(d2+d4+d6) doit se terminer par d7.
func DeaChecksum(value string) bool {
	if len(value) < 9 {
		return false
	}
	digits := value[2:]
	if len(digits) != 7 {
		return false
	}
	sum := 0
	for i, c := range digits[:6] {
		if c < '0' || c > '9' {
			return false
		}
		d := int(c - '0')
		if i%2 == 1 {
			d *= 2
		}
		sum += d
	}
	check := digits[6]
	if check < '0' || check > '9' {
		return false
	}
	return sum%10 == int(check-'0')
}
