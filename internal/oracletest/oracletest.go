// Package oracletest exécute le corpus oracle (internal/testdata/oracle.jsonl)
// contre un ensemble de recognizers. Seuls les cas portant sur des entités
// supportées par l'ensemble sont évalués.
package oracletest

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"github.com/YoLaub/PresidioGo/pii"
	"github.com/YoLaub/PresidioGo/recognizer"
)

// Expected est une entité attendue d'un cas oracle.
type Expected struct {
	EntityType string  `json:"entity_type"`
	Start      int     `json:"start"`
	End        int     `json:"end"`
	MinScore   float64 `json:"min_score"`
}

// Case est une ligne du corpus.
type Case struct {
	ID       string     `json:"id"`
	Text     string     `json:"text"`
	Expected []Expected `json:"expected"`
	Forbid   []string   `json:"forbid"`
}

// Load lit le corpus oracle complet.
func Load(t *testing.T) []Case {
	t.Helper()
	_, self, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(self), "..", "testdata", "oracle.jsonl")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("ouverture oracle.jsonl : %v", err)
	}
	defer f.Close()

	var cases []Case
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if len(sc.Bytes()) == 0 {
			continue
		}
		var c Case
		if err := json.Unmarshal(sc.Bytes(), &c); err != nil {
			t.Fatalf("ligne oracle invalide : %v", err)
		}
		cases = append(cases, c)
	}
	return cases
}

// Run évalue les recognizers sur chaque cas oracle pertinent : les entités
// attendues (des types supportés) doivent être trouvées aux offsets runes
// exacts avec un score suffisant ; les entités interdites ne doivent pas
// apparaître.
func Run(t *testing.T, recognizers []recognizer.Recognizer) {
	t.Helper()
	var supported []string
	for _, rec := range recognizers {
		supported = append(supported, rec.SupportedEntities()...)
	}

	for _, c := range Load(t) {
		expected := filterExpected(c.Expected, supported)
		forbid := filterStrings(c.Forbid, supported)
		if len(expected) == 0 && len(forbid) == 0 {
			continue
		}
		t.Run(c.ID, func(t *testing.T) {
			var found []pii.Result
			for _, rec := range recognizers {
				results, err := rec.Analyze(context.Background(), c.Text, nil)
				if err != nil {
					t.Fatalf("%s : %v", rec.Name(), err)
				}
				found = append(found, results...)
			}
			for _, exp := range expected {
				if !slices.ContainsFunc(found, func(r pii.Result) bool {
					return r.EntityType == exp.EntityType &&
						r.Start == exp.Start && r.End == exp.End &&
						r.Score >= exp.MinScore
				}) {
					t.Errorf("attendu %s [%d:%d] score>=%.2f — trouvé : %+v",
						exp.EntityType, exp.Start, exp.End, exp.MinScore, found)
				}
			}
			for _, forbidden := range forbid {
				for _, r := range found {
					if r.EntityType == forbidden {
						t.Errorf("entité interdite %s détectée : %+v", forbidden, r)
					}
				}
			}
		})
	}
}

func filterExpected(in []Expected, supported []string) []Expected {
	var out []Expected
	for _, e := range in {
		if slices.Contains(supported, e.EntityType) {
			out = append(out, e)
		}
	}
	return out
}

func filterStrings(in, supported []string) []string {
	var out []string
	for _, s := range in {
		if slices.Contains(supported, s) {
			out = append(out, s)
		}
	}
	return out
}
