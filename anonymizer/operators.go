// Package anonymizer remplace les entités PII détectées par le résultat
// d'opérateurs (replace, mask, hash, encrypt…), avec résolution des
// chevauchements entre entités.
package anonymizer

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
)

// Operator transforme la valeur d'une entité détectée.
type Operator interface {
	Name() string
	Operate(value string) (string, error)
}

type funcOperator struct {
	name string
	fn   func(string) (string, error)
}

func (o funcOperator) Name() string                         { return o.name }
func (o funcOperator) Operate(value string) (string, error) { return o.fn(value) }

// Replace remplace l'entité par une valeur fixe.
func Replace(newValue string) Operator {
	return funcOperator{"replace", func(string) (string, error) { return newValue, nil }}
}

// Redact supprime l'entité du texte.
func Redact() Operator {
	return funcOperator{"redact", func(string) (string, error) { return "", nil }}
}

// Keep laisse l'entité intacte (mais la trace dans les items).
func Keep() Operator {
	return funcOperator{"keep", func(v string) (string, error) { return v, nil }}
}

// Custom applique une fonction fournie par l'appelant.
func Custom(name string, fn func(string) (string, error)) Operator {
	return funcOperator{name, fn}
}

// Mask masque charsToMask caractères (runes) avec maskChar, depuis le début
// ou depuis la fin. Si charsToMask dépasse la longueur, tout est masqué.
func Mask(maskChar rune, charsToMask int, fromEnd bool) Operator {
	return funcOperator{"mask", func(v string) (string, error) {
		runes := []rune(v)
		n := min(charsToMask, len(runes))
		start := 0
		if fromEnd {
			start = len(runes) - n
		}
		for i := start; i < start+n; i++ {
			runes[i] = maskChar
		}
		return string(runes), nil
	}}
}

// Hash remplace l'entité par son SHA-256 en hexadécimal (déterministe).
func Hash() Operator {
	return funcOperator{"hash", func(v string) (string, error) {
		sum := sha256.Sum256([]byte(v))
		return hex.EncodeToString(sum[:]), nil
	}}
}

func newGCM(key []byte) (cipher.AEAD, error) {
	switch len(key) {
	case 16, 24, 32:
	default:
		return nil, errors.New("anonymizer: la clé AES doit faire 16, 24 ou 32 octets")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// Encrypt chiffre l'entité en AES-GCM ; sortie base64(nonce||ciphertext).
// Réversible via Decrypt avec la même clé.
func Encrypt(key []byte) (Operator, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}
	return funcOperator{"encrypt", func(v string) (string, error) {
		nonce := make([]byte, gcm.NonceSize())
		if _, err := rand.Read(nonce); err != nil {
			return "", err
		}
		sealed := gcm.Seal(nonce, nonce, []byte(v), nil)
		return base64.StdEncoding.EncodeToString(sealed), nil
	}}, nil
}

// Decrypt inverse Encrypt.
func Decrypt(key []byte) (Operator, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}
	return funcOperator{"decrypt", func(v string) (string, error) {
		raw, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return "", fmt.Errorf("anonymizer: base64 invalide : %w", err)
		}
		if len(raw) < gcm.NonceSize() {
			return "", errors.New("anonymizer: données chiffrées tronquées")
		}
		plain, err := gcm.Open(nil, raw[:gcm.NonceSize()], raw[gcm.NonceSize():], nil)
		if err != nil {
			return "", fmt.Errorf("anonymizer: déchiffrement impossible : %w", err)
		}
		return string(plain), nil
	}}, nil
}
