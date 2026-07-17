// Package ner expose les entités nommées du moteur NLP (nlp.Artifacts)
// comme résultats PII — équivalent des ner recognizers du fork. Il ne fait
// aucune inférence lui-même : il lit ce que le moteur (ONNX ou autre) a
// produit, et mappe les labels CoNLL vers les entités Presidio.
package ner

import (
	"context"

	"github.com/YoLaub/presidigo-go/nlp"
	"github.com/YoLaub/presidigo-go/pii"
)

// DefaultMapping mappe les labels CoNLL du modèle vers les entités Presidio
// (même correspondance que la configuration transformers du fork).
var DefaultMapping = map[string]string{
	"PER":  "PERSON",
	"LOC":  "LOCATION",
	"ORG":  "ORGANIZATION",
	"MISC": "NRP",
}

// Recognizer lit les NerEntities des artifacts NLP.
type Recognizer struct {
	language string
	mapping  map[string]string
	entities []string
}

// New crée le recognizer NER pour une langue, avec le mapping par défaut.
func New(language string) *Recognizer {
	return NewWithMapping(language, DefaultMapping)
}

// NewWithMapping crée le recognizer avec un mapping label→entité personnalisé.
func NewWithMapping(language string, mapping map[string]string) *Recognizer {
	entities := make([]string, 0, len(mapping))
	seen := map[string]bool{}
	for _, e := range mapping {
		if !seen[e] {
			entities = append(entities, e)
			seen[e] = true
		}
	}
	return &Recognizer{language: language, mapping: mapping, entities: entities}
}

func (r *Recognizer) Name() string                { return "NerRecognizer" }
func (r *Recognizer) SupportedEntities() []string { return r.entities }
func (r *Recognizer) Language() string            { return r.language }

// Analyze convertit les entités NER des artifacts en résultats PII.
// Les labels absents du mapping sont ignorés.
func (r *Recognizer) Analyze(_ context.Context, _ string, artifacts *nlp.Artifacts) ([]pii.Result, error) {
	if artifacts == nil {
		return nil, nil
	}
	var results []pii.Result
	for _, e := range artifacts.NerEntities {
		entity, ok := r.mapping[e.Label]
		if !ok {
			continue
		}
		results = append(results, pii.Result{
			EntityType: entity,
			Start:      e.Start,
			End:        e.End,
			Score:      e.Score,
			Explanation: &pii.Explanation{
				Recognizer:    r.Name(),
				Pattern:       "ner:" + e.Label,
				OriginalScore: e.Score,
			},
		})
	}
	return results, nil
}
