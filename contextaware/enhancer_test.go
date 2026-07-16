package contextaware_test

import (
	"context"
	"math"
	"regexp"
	"testing"

	"github.com/YoLaub/presidigo-go/contextaware"
	"github.com/YoLaub/presidigo-go/pii"
	"github.com/YoLaub/presidigo-go/recognizer"
)

func recWithContext(t *testing.T, score float64, words ...string) *recognizer.PatternRecognizer {
	t.Helper()
	r, err := recognizer.NewPattern("PhoneRecognizer", "PHONE_NUMBER", "fr",
		[]pii.Pattern{{
			Name:  "digits10",
			Regex: regexp.MustCompile(`\b0\d{9}\b`),
			Score: score,
		}},
		recognizer.WithContextWords(words...))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func analyze(t *testing.T, rec recognizer.Recognizer, text string) []pii.Result {
	t.Helper()
	results, err := rec.Analyze(context.Background(), text, nil)
	if err != nil {
		t.Fatal(err)
	}
	return results
}

func almost(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestEnhance_BoostQuandMotDeContexteAvant(t *testing.T) {
	rec := recWithContext(t, 0.5, "téléphone", "tel")
	text := "mon numéro de téléphone est 0601020304"
	results := analyze(t, rec, text)

	enh := contextaware.New()
	out := enh.Enhance(text, results, []recognizer.Recognizer{rec})
	if len(out) != 1 || !almost(out[0].Score, 0.85) {
		t.Fatalf("attendu score 0.85 (0.5+0.35), obtenu %+v", out)
	}
	if out[0].Explanation == nil || !almost(out[0].Explanation.ContextBoost, 0.35) {
		t.Errorf("ContextBoost non tracé : %+v", out[0].Explanation)
	}
}

func TestEnhance_PasDeBoostSansMot(t *testing.T) {
	rec := recWithContext(t, 0.5, "téléphone")
	text := "voici la valeur 0601020304"
	out := contextaware.New().Enhance(text, analyze(t, rec, text), []recognizer.Recognizer{rec})
	if len(out) != 1 || !almost(out[0].Score, 0.5) {
		t.Fatalf("aucun boost attendu, obtenu %+v", out)
	}
}

func TestEnhance_MotTropLoinHorsFenetre(t *testing.T) {
	// « téléphone » est à plus de 5 mots de l'entité : hors fenêtre (préfixe 5).
	rec := recWithContext(t, 0.5, "téléphone")
	text := "téléphone un deux trois quatre cinq six 0601020304"
	out := contextaware.New().Enhance(text, analyze(t, rec, text), []recognizer.Recognizer{rec})
	if len(out) != 1 || !almost(out[0].Score, 0.5) {
		t.Fatalf("mot hors fenêtre : aucun boost attendu, obtenu %+v", out)
	}
}

func TestEnhance_ScoreMinimumAvecContexte(t *testing.T) {
	// 0.01 + 0.35 = 0.36 < 0.4 : remonté au minimum (fork : min_score_with_context_similarity).
	rec := recWithContext(t, 0.01, "tel")
	text := "tel 0601020304"
	out := contextaware.New().Enhance(text, analyze(t, rec, text), []recognizer.Recognizer{rec})
	if len(out) != 1 || !almost(out[0].Score, 0.4) {
		t.Fatalf("attendu 0.4, obtenu %+v", out)
	}
}

func TestEnhance_PlafonneAMaxScore(t *testing.T) {
	rec := recWithContext(t, 0.9, "tel")
	text := "tel 0601020304"
	out := contextaware.New().Enhance(text, analyze(t, rec, text), []recognizer.Recognizer{rec})
	if len(out) != 1 || !almost(out[0].Score, pii.MaxScore) {
		t.Fatalf("attendu plafond %v, obtenu %+v", pii.MaxScore, out)
	}
}

func TestEnhance_RecognizerSansContexteInchange(t *testing.T) {
	r, err := recognizer.NewPattern("NoCtx", "PHONE_NUMBER", "fr",
		[]pii.Pattern{{Name: "d", Regex: regexp.MustCompile(`\b0\d{9}\b`), Score: 0.5}})
	if err != nil {
		t.Fatal(err)
	}
	text := "téléphone 0601020304"
	out := contextaware.New().Enhance(text, analyze(t, r, text), []recognizer.Recognizer{r})
	if len(out) != 1 || !almost(out[0].Score, 0.5) {
		t.Fatalf("recognizer sans mots de contexte : inchangé attendu, obtenu %+v", out)
	}
}

func TestContextWords_ExposeParLeRecognizer(t *testing.T) {
	rec := recWithContext(t, 0.5, "tel", "téléphone")
	words := rec.ContextWords()
	if len(words) != 2 {
		t.Fatalf("ContextWords = %v", words)
	}
}
