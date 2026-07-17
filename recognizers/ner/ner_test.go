package ner_test

import (
	"context"
	"testing"

	"github.com/YoLaub/presidigo-go/nlp"
	"github.com/YoLaub/presidigo-go/recognizers/ner"
)

func TestAnalyze_MappeLesLabelsVersLesEntitesPresidio(t *testing.T) {
	rec := ner.New("en")
	artifacts := &nlp.Artifacts{NerEntities: []nlp.NerEntity{
		{Label: "PER", Start: 0, End: 7, Score: 0.97},
		{Label: "LOC", Start: 17, End: 22, Score: 0.90},
		{Label: "ORG", Start: 30, End: 34, Score: 0.88},
		{Label: "MISC", Start: 40, End: 45, Score: 0.80},
	}}
	results, err := rec.Analyze(context.Background(), "Johnson works in Paris chez ACME etc.", artifacts)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"PERSON": false, "LOCATION": false, "ORGANIZATION": false, "NRP": false}
	for _, r := range results {
		if _, ok := want[r.EntityType]; !ok {
			t.Errorf("entité inattendue %q", r.EntityType)
		}
		want[r.EntityType] = true
	}
	for k, seen := range want {
		if !seen {
			t.Errorf("entité %s manquante : %+v", k, results)
		}
	}
	// Scores et offsets transmis tels quels.
	if results[0].Start != 0 || results[0].End != 7 || results[0].Score != 0.97 {
		t.Errorf("résultat PER = %+v", results[0])
	}
	if results[0].Explanation == nil || results[0].Explanation.Recognizer != "NerRecognizer" {
		t.Errorf("Explanation = %+v", results[0].Explanation)
	}
}

func TestAnalyze_SansArtifactsRetourneRien(t *testing.T) {
	rec := ner.New("en")
	results, err := rec.Analyze(context.Background(), "Johnson", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("sans artifacts : aucun résultat attendu, obtenu %+v", results)
	}
}

func TestAnalyze_LabelInconnuIgnore(t *testing.T) {
	rec := ner.New("en")
	artifacts := &nlp.Artifacts{NerEntities: []nlp.NerEntity{{Label: "XYZ", Start: 0, End: 3, Score: 0.9}}}
	results, err := rec.Analyze(context.Background(), "abc", artifacts)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("label inconnu : ignoré attendu, obtenu %+v", results)
	}
}

func TestSupportedEntities(t *testing.T) {
	got := ner.New("en").SupportedEntities()
	if len(got) != 4 {
		t.Errorf("SupportedEntities = %v", got)
	}
}
