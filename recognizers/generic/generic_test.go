package generic_test

import (
	"context"
	"testing"

	"github.com/YoLaub/presidigo-go/internal/oracletest"
	"github.com/YoLaub/presidigo-go/pii"
	"github.com/YoLaub/presidigo-go/recognizers/generic"
)

// TestOracle exécute les recognizers génériques sur le corpus oracle partagé.
func TestOracle(t *testing.T) {
	recognizers := generic.All("en")
	if len(recognizers) < 7 {
		t.Fatalf("generic.All doit fournir au moins 7 recognizers, obtenu %d", len(recognizers))
	}
	oracletest.Run(t, recognizers)
}

func TestLuhn(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want bool
	}{
		{"4012888888881881", true},
		{"5555555555554444", true},
		{"4012888888881882", false},
	} {
		if got := generic.Luhn(tc.in); got != tc.want {
			t.Errorf("Luhn(%q) = %v, attendu %v", tc.in, got, tc.want)
		}
	}
}

func TestIbanMod97(t *testing.T) {
	if !generic.IbanMod97("GB29NWBK60161331926819") {
		t.Error("GB29NWBK60161331926819 devrait être valide (mod 97)")
	}
	if generic.IbanMod97("GB29NWBK60161331926820") {
		t.Error("GB29NWBK60161331926820 devrait être invalide")
	}
}

func TestCreditCard_ExclusionTreizeChiffresCommencantPar1(t *testing.T) {
	// Le fork Python exclut « 1 suivi d'exactement 12 chiffres » par lookahead
	// négatif (non supporté par RE2) : porté en validation Go.
	rec := generic.NewCreditCard("en")
	results, err := rec.Analyze(context.Background(), "1000000000009", nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range results {
		t.Errorf("13 chiffres commençant par 1 ne doit pas matcher : %+v", r)
	}
}

func TestCrypto_Bech32Valide(t *testing.T) {
	rec := generic.NewCrypto("en")
	text := "bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq"
	results, err := rec.Analyze(context.Background(), text, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Score != pii.MaxScore {
		t.Errorf("adresse bech32 valide attendue au score max, obtenu %+v", results)
	}
}
