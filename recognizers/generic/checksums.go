package generic

// Luhn vérifie la somme de contrôle Luhn d'une chaîne de chiffres
// (cartes bancaires, SIREN/SIRET…).
func Luhn(digits string) bool {
	sum := 0
	double := false
	for i := len(digits) - 1; i >= 0; i-- {
		c := digits[i]
		if c < '0' || c > '9' {
			return false
		}
		d := int(c - '0')
		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}
	return sum%10 == 0
}

// IbanMod97 vérifie la somme de contrôle ISO 7064 mod 97-10 d'un IBAN
// (sans espaces) : les 4 premiers caractères passent à la fin, les lettres
// valent 10..35, et le reste modulo 97 doit être 1.
func IbanMod97(iban string) bool {
	if len(iban) < 5 {
		return false
	}
	rearranged := iban[4:] + iban[:4]
	n := 0
	for i := 0; i < len(rearranged); i++ {
		c := rearranged[i]
		switch {
		case c >= '0' && c <= '9':
			n = (n*10 + int(c-'0')) % 97
		case c >= 'A' && c <= 'Z':
			v := int(c-'A') + 10
			n = (n*100 + v) % 97
		default:
			return false
		}
	}
	return n == 1
}
