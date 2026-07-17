package recognizer_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/YoLaub/PresidioGo/pii"
	"github.com/YoLaub/PresidioGo/recognizer"
)

func emailRecognizer(t *testing.T, opts ...recognizer.Option) *recognizer.PatternRecognizer {
	t.Helper()
	r, err := recognizer.NewPattern("EmailRecognizer", "EMAIL_ADDRESS", "en",
		[]pii.Pattern{{
			Name:  "email-basic",
			Regex: regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`),
			Score: 0.6,
		}}, opts...)
	if err != nil {
		t.Fatalf("NewPattern: %v", err)
	}
	return r
}

func TestPatternRecognizer_DetecteAvecOffsetsRunes(t *testing.T) {
	r := emailRecognizer(t)

	// « Prénom » et « José » contiennent des caractères multi-bytes : les offsets
	// attendus sont en RUNES (cas oracle email-rune-offsets).
	text := "Prénom : José — email : info@presidio.site"
	results, err := r.Analyze(context.Background(), text, nil)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("attendu 1 résultat, obtenu %d : %+v", len(results), results)
	}
	got := results[0]
	if got.EntityType != "EMAIL_ADDRESS" {
		t.Errorf("EntityType = %q, attendu EMAIL_ADDRESS", got.EntityType)
	}
	if got.Start != 24 || got.End != 42 {
		t.Errorf("offsets = (%d,%d), attendu (24,42) en runes", got.Start, got.End)
	}
	if got.Score != 0.6 {
		t.Errorf("Score = %v, attendu 0.6", got.Score)
	}
	if extracted := string([]rune(text)[got.Start:got.End]); extracted != "info@presidio.site" {
		t.Errorf("extraction par offsets runes = %q", extracted)
	}
}

func TestPatternRecognizer_TexteVide(t *testing.T) {
	r := emailRecognizer(t)
	results, err := r.Analyze(context.Background(), "", nil)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("attendu 0 résultat sur texte vide, obtenu %d", len(results))
	}
}

func TestPatternRecognizer_ValidationRejette(t *testing.T) {
	// Validate → false : le match est rejeté (ex. checksum Luhn KO).
	reject := func(_ string) *bool { v := false; return &v }
	r := emailRecognizer(t, recognizer.WithValidate(reject))

	results, err := r.Analyze(context.Background(), "mail: info@presidio.site", nil)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("attendu 0 résultat (invalidé), obtenu %d", len(results))
	}
}

func TestPatternRecognizer_ValidationConfirme(t *testing.T) {
	// Validate → true : le score monte au maximum.
	confirm := func(_ string) *bool { v := true; return &v }
	r := emailRecognizer(t, recognizer.WithValidate(confirm))

	results, err := r.Analyze(context.Background(), "mail: info@presidio.site", nil)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("attendu 1 résultat, obtenu %d", len(results))
	}
	if results[0].Score != pii.MaxScore {
		t.Errorf("Score = %v, attendu pii.MaxScore (%v)", results[0].Score, pii.MaxScore)
	}
}

func TestPatternRecognizer_ValidationNeutre(t *testing.T) {
	// Validate → nil : le score du pattern est conservé.
	neutral := func(_ string) *bool { return nil }
	r := emailRecognizer(t, recognizer.WithValidate(neutral))

	results, err := r.Analyze(context.Background(), "mail: info@presidio.site", nil)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if len(results) != 1 || results[0].Score != 0.6 {
		t.Fatalf("attendu 1 résultat au score 0.6, obtenu %+v", results)
	}
}

func TestPatternRecognizer_ExplicationRenseignee(t *testing.T) {
	r := emailRecognizer(t)
	results, _ := r.Analyze(context.Background(), "info@presidio.site", nil)
	if len(results) != 1 || results[0].Explanation == nil {
		t.Fatalf("Explanation manquante : %+v", results)
	}
	e := results[0].Explanation
	if e.Recognizer != "EmailRecognizer" || e.Pattern != "email-basic" {
		t.Errorf("Explanation = %+v", e)
	}
}

func TestPatternRecognizer_Metadonnees(t *testing.T) {
	r := emailRecognizer(t)
	if r.Name() != "EmailRecognizer" {
		t.Errorf("Name = %q", r.Name())
	}
	if got := r.SupportedEntities(); len(got) != 1 || got[0] != "EMAIL_ADDRESS" {
		t.Errorf("SupportedEntities = %v", got)
	}
	if r.Language() != "en" {
		t.Errorf("Language = %q", r.Language())
	}
}

func TestNewPattern_ErreurSansPattern(t *testing.T) {
	_, err := recognizer.NewPattern("X", "Y", "en", nil)
	if err == nil {
		t.Fatal("attendu une erreur quand aucun pattern n'est fourni")
	}
}
