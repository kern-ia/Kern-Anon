package generic

import (
	"bytes"
	"crypto/sha256"
	"math/big"
	"regexp"
	"strings"

	"github.com/YoLaub/PresidioGo/pii"
	"github.com/YoLaub/PresidioGo/recognizer"
)

var cryptoPattern = pii.Pattern{
	Name:  "Crypto (Medium)",
	Regex: regexp.MustCompile(`(bc1|[13])[a-zA-HJ-NP-Z0-9]{25,59}`),
	Score: 0.5,
}

// NewCrypto détecte les adresses Bitcoin (CRYPTO) : P2PKH/P2SH validées par
// base58check (double SHA-256), bech32/bech32m validées par polymod.
func NewCrypto(language string) *recognizer.PatternRecognizer {
	return mustPattern("CryptoRecognizer", "CRYPTO", language,
		[]pii.Pattern{cryptoPattern},
		recognizer.WithContextWords("wallet", "btc", "bitcoin", "crypto"),
		recognizer.WithValidate(func(match string) *bool {
			var ok bool
			switch {
			case strings.HasPrefix(match, "bc1"):
				ok = bech32Verify(match)
			case match[0] == '1' || match[0] == '3':
				ok = base58CheckVerify(match)
			}
			return &ok
		}))
}

const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// base58CheckVerify décode l'adresse en base58 et vérifie que les 4 derniers
// octets égalent le double SHA-256 du reste.
func base58CheckVerify(addr string) bool {
	n := new(big.Int)
	radix := big.NewInt(58)
	for i := 0; i < len(addr); i++ {
		idx := strings.IndexByte(base58Alphabet, addr[i])
		if idx < 0 {
			return false
		}
		n.Mul(n, radix)
		n.Add(n, big.NewInt(int64(idx)))
	}
	decoded := n.Bytes()
	// Chaque '1' de tête encode un octet nul.
	for i := 0; i < len(addr) && addr[i] == '1'; i++ {
		decoded = append([]byte{0}, decoded...)
	}
	if len(decoded) < 5 {
		return false
	}
	payload, checksum := decoded[:len(decoded)-4], decoded[len(decoded)-4:]
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	return bytes.Equal(checksum, second[:4])
}

const bech32Charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

// bech32Verify vérifie la somme de contrôle bech32 (BIP-173) ou bech32m
// (BIP-350) d'une adresse.
func bech32Verify(addr string) bool {
	if strings.ToLower(addr) != addr && strings.ToUpper(addr) != addr {
		return false // casse mélangée interdite
	}
	addr = strings.ToLower(addr)
	sep := strings.LastIndexByte(addr, '1')
	if sep < 1 || sep+7 > len(addr) {
		return false
	}
	hrp, data := addr[:sep], addr[sep+1:]
	values := make([]int, 0, len(hrp)*2+1+len(data))
	for i := 0; i < len(hrp); i++ {
		values = append(values, int(hrp[i])>>5)
	}
	values = append(values, 0)
	for i := 0; i < len(hrp); i++ {
		values = append(values, int(hrp[i])&31)
	}
	for i := 0; i < len(data); i++ {
		v := strings.IndexByte(bech32Charset, data[i])
		if v < 0 {
			return false
		}
		values = append(values, v)
	}
	chk := bech32Polymod(values)
	return chk == 1 || chk == 0x2bc830a3
}

func bech32Polymod(values []int) int {
	gen := [5]int{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}
	chk := 1
	for _, v := range values {
		top := chk >> 25
		chk = (chk&0x1ffffff)<<5 ^ v
		for i := 0; i < 5; i++ {
			if (top>>i)&1 == 1 {
				chk ^= gen[i]
			}
		}
	}
	return chk
}
