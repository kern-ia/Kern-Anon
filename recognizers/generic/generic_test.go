package generic_test

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"slices"
	"testing"

	"github.com/YoLaub/presidigo-go/pii"
	"github.com/YoLaub/presidigo-go/recognizers/generic"
)

type oracleExpected struct {
	EntityType string  `json:"entity_type"`
	Start      int     `json:"start"`
	End        int     `json:"end"`
	MinScore   float64 `json:"min_score"`
}

type oracleCase struct {
	ID       string           `json:"id"`
	Text     string           `json:"text"`
	Expected []oracleExpected `json:"expected"`
	Forbid   []string         `json:"forbid"`
}

func loadOracle(t *testing.T) []oracleCase {
	t.Helper()
	f, err := os.Open("../../internal/testdata/oracle.jsonl")
	if err != nil {
		t.Fatalf("ouverture oracle.jsonl : %v", err)
	}
	defer f.Close()

	var cases []oracleCase
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if len(sc.Bytes()) == 0 {
			continue
		}
		var c oracleCase
		if err := json.Unmarshal(sc.Bytes(), &c); err != nil {
			t.Fatalf("ligne oracle invalide : %v", err)
		}
		cases = append(cases, c)
	}
	return cases
}

// TestOracle exécute tous les recognizers génériques sur chaque cas du corpus
// oracle : les entités attendues doivent être trouvées aux offsets runes exacts
// avec un score suffisant ; les entités interdites ne doivent pas apparaître.
func TestOracle(t *testing.T) {
	recognizers := generic.All("en")
	if len(recognizers) < 7 {
		t.Fatalf("generic.All doit fournir au moins 7 recognizers, obtenu %d", len(recognizers))
	}

	for _, c := range loadOracle(t) {
		t.Run(c.ID, func(t *testing.T) {
			var found []pii.Result
			for _, rec := range recognizers {
				results, err := rec.Analyze(context.Background(), c.Text, nil)
				if err != nil {
					t.Fatalf("%s : %v", rec.Name(), err)
				}
				found = append(found, results...)
			}
			for _, exp := range c.Expected {
				if !slices.ContainsFunc(found, func(r pii.Result) bool {
					return r.EntityType == exp.EntityType &&
						r.Start == exp.Start && r.End == exp.End &&
						r.Score >= exp.MinScore
				}) {
					t.Errorf("attendu %s [%d:%d] score>=%.2f — trouvé : %+v",
						exp.EntityType, exp.Start, exp.End, exp.MinScore, found)
				}
			}
			for _, forbidden := range c.Forbid {
				for _, r := range found {
					if r.EntityType == forbidden {
						t.Errorf("entité interdite %s détectée : %+v", forbidden, r)
					}
				}
			}
		})
	}
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
