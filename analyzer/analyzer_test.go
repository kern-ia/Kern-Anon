package analyzer_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/YoLaub/presidigo-go/analyzer"
	"github.com/YoLaub/presidigo-go/nlp"
	"github.com/YoLaub/presidigo-go/pii"
	"github.com/YoLaub/presidigo-go/recognizer"
	"github.com/YoLaub/presidigo-go/recognizers/generic"
	"github.com/YoLaub/presidigo-go/registry"
)

func genericRegistry(langs ...string) *registry.Registry {
	reg := registry.New()
	for _, lang := range langs {
		for _, rec := range generic.All(lang) {
			reg.Add(rec)
		}
	}
	return reg
}

func TestAnalyze_PipelineComplet(t *testing.T) {
	eng, err := analyzer.New(analyzer.WithRegistry(genericRegistry("fr")))
	if err != nil {
		t.Fatal(err)
	}
	// « email » précède l'entité : le boost contextuel doit s'appliquer (0.5+0.35).
	results, err := eng.Analyze(context.Background(),
		"Prénom : José — email : info@presidio.site", analyzer.Language("fr"))
	if err != nil {
		t.Fatal(err)
	}
	var email *pii.Result
	for i := range results {
		if results[i].EntityType == "EMAIL_ADDRESS" {
			email = &results[i]
		}
	}
	if email == nil {
		t.Fatalf("EMAIL_ADDRESS non détecté : %+v", results)
	}
	if email.Start != 24 || email.End != 42 {
		t.Errorf("offsets = (%d,%d)", email.Start, email.End)
	}
	if email.Score < 0.84 || email.Score > 0.86 {
		t.Errorf("boost contextuel attendu (≈0.85), obtenu %v", email.Score)
	}
}

func TestAnalyze_FiltreMinScore(t *testing.T) {
	eng, _ := analyzer.New(analyzer.WithRegistry(genericRegistry("en")))
	// Sans mot de contexte : URL score 0.5, IP 0.6. MinScore(0.6) écarte l'URL.
	results, err := eng.Analyze(context.Background(),
		"microsoft.com 192.168.0.1", analyzer.MinScore(0.6))
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range results {
		if r.Score < 0.6 {
			t.Errorf("résultat sous le seuil : %+v", r)
		}
		if r.EntityType == "URL" {
			t.Errorf("URL (0.5) aurait dû être filtrée : %+v", r)
		}
	}
}

func TestAnalyze_FiltreEntites(t *testing.T) {
	eng, _ := analyzer.New(analyzer.WithRegistry(genericRegistry("en")))
	results, err := eng.Analyze(context.Background(),
		"card 4012888888881881 mail info@presidio.site",
		analyzer.Entities("CREDIT_CARD"))
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].EntityType != "CREDIT_CARD" {
		t.Fatalf("seul CREDIT_CARD attendu, obtenu %+v", results)
	}
}

func TestAnalyze_LangueParDefautEN(t *testing.T) {
	eng, _ := analyzer.New(analyzer.WithRegistry(genericRegistry("en", "fr")))
	results, err := eng.Analyze(context.Background(), "mail info@presidio.site")
	if err != nil {
		t.Fatal(err)
	}
	// Registry en+fr : sans option Language, seuls les recognizers "en"
	// tournent — un seul EMAIL_ADDRESS (pas de doublon fr).
	count := 0
	for _, r := range results {
		if r.EntityType == "EMAIL_ADDRESS" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("attendu 1 EMAIL_ADDRESS (langue par défaut en), obtenu %d", count)
	}
}

func TestAnalyze_DedoublonneMemeSpanMemeEntite(t *testing.T) {
	// Deux recognizers produisent la même entité : le span contenu dans un
	// span plus large de même type et de score inférieur est écarté.
	strong, _ := recognizer.NewPattern("Strong", "PHONE_NUMBER", "en",
		[]pii.Pattern{{Name: "full", Regex: regexp.MustCompile(`\b0\d{9}\b`), Score: 0.7}})
	weak, _ := recognizer.NewPattern("Weak", "PHONE_NUMBER", "en",
		[]pii.Pattern{{Name: "partial", Regex: regexp.MustCompile(`\d{4}\b`), Score: 0.3}})

	reg := registry.New()
	reg.Add(strong)
	reg.Add(weak)
	eng, _ := analyzer.New(analyzer.WithRegistry(reg))

	results, err := eng.Analyze(context.Background(), "tel 0601020304")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Explanation.Recognizer != "Strong" {
		t.Fatalf("le span contenu (Weak) devait être écarté : %+v", results)
	}
}

func TestAnalyze_ResultatsTriesParPosition(t *testing.T) {
	eng, _ := analyzer.New(analyzer.WithRegistry(genericRegistry("en")))
	results, err := eng.Analyze(context.Background(),
		"ip 192.168.0.1 puis carte 4012888888881881")
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < len(results); i++ {
		if results[i].Start < results[i-1].Start {
			t.Fatalf("résultats non triés : %+v", results)
		}
	}
}

func TestNew_RegistryObligatoire(t *testing.T) {
	if _, err := analyzer.New(); err == nil {
		t.Fatal("erreur attendue sans registry")
	}
}

// fakeNlpRecognizer capture les artifacts reçus pour vérifier le câblage NLP.
type fakeNlpRecognizer struct {
	got *nlp.Artifacts
}

func (f *fakeNlpRecognizer) Name() string                { return "FakeNlp" }
func (f *fakeNlpRecognizer) SupportedEntities() []string { return []string{"X"} }
func (f *fakeNlpRecognizer) Language() string            { return "en" }
func (f *fakeNlpRecognizer) Analyze(_ context.Context, _ string, a *nlp.Artifacts) ([]pii.Result, error) {
	f.got = a
	return nil, nil
}

func TestAnalyze_ArtifactsTransmisAuxRecognizers(t *testing.T) {
	fake := &fakeNlpRecognizer{}
	reg := registry.New()
	reg.Add(fake)
	eng, _ := analyzer.New(
		analyzer.WithRegistry(reg),
		analyzer.WithNlpEngine(nlp.NoOp{}),
	)
	if _, err := eng.Analyze(context.Background(), "bonjour"); err != nil {
		t.Fatal(err)
	}
	if fake.got == nil {
		t.Fatal("les artifacts du moteur NLP doivent être transmis aux recognizers")
	}
}
