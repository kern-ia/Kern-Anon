// Package anonymizeoracle exécute le corpus internal/testdata/anonymize.jsonl
// à travers le pipeline complet analyzer → anonymizer et vérifie le texte
// anonymisé produit. Contrairement à internal/oracletest (recognizers seuls),
// ce corpus verrouille le comportement bout en bout : détection, résolution
// des chevauchements et substitution.
package anonymizeoracle

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/YoLaub/PresidioGo/analyzer"
	"github.com/YoLaub/PresidioGo/anonymizer"
	"github.com/YoLaub/PresidioGo/registry"
)

// Case est une ligne du corpus.
type Case struct {
	ID           string   `json:"id"`
	Language     string   `json:"language"`
	Text         string   `json:"text"`
	Entities     []string `json:"entities"`
	ExpectedText string   `json:"expected_text"`
}

// Load lit le corpus anonymize.jsonl.
func Load(t *testing.T) []Case {
	t.Helper()
	_, self, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(self), "..", "testdata", "anonymize.jsonl")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("ouverture anonymize.jsonl : %v", err)
	}
	defer func() { _ = f.Close() }()

	var cases []Case
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if len(sc.Bytes()) == 0 {
			continue
		}
		var c Case
		if err := json.Unmarshal(sc.Bytes(), &c); err != nil {
			t.Fatalf("ligne anonymize.jsonl invalide : %v", err)
		}
		cases = append(cases, c)
	}
	return cases
}

// Run rejoue chaque cas via l'API publique (analyzer.New + registry.Default
// + anonymizer.New, opérateurs par défaut) et compare le texte obtenu au
// texte attendu.
func Run(t *testing.T) {
	t.Helper()
	for _, c := range Load(t) {
		t.Run(c.ID, func(t *testing.T) {
			eng, err := analyzer.New(
				analyzer.WithRegistry(registry.Default(c.Language)),
				analyzer.WithDefaultLanguage(c.Language),
			)
			if err != nil {
				t.Fatalf("analyzer.New : %v", err)
			}
			results, err := eng.Analyze(context.Background(), c.Text, analyzer.MinScore(0.4))
			if err != nil {
				t.Fatalf("Analyze : %v", err)
			}
			out, err := anonymizer.New().Anonymize(c.Text, results, nil)
			if err != nil {
				t.Fatalf("Anonymize : %v", err)
			}
			if out.Text != c.ExpectedText {
				t.Errorf("texte anonymisé =\n  %q\nattendu :\n  %q", out.Text, c.ExpectedText)
			}
		})
	}
}
