// Package nlp définit les artefacts issus du traitement NLP d'un texte,
// partagés entre tous les recognizers. L'interface du moteur (NlpEngine)
// et son implémentation ONNX arrivent avec la feature nlp-onnx.
package nlp

import "context"

// NerEntity est une entité nommée produite par le moteur NLP,
// offsets en runes.
type NerEntity struct {
	Label string // "PER", "LOC", "ORG", "MISC", …
	Start int
	End   int
	Score float64
}

// Artifacts porte les résultats du traitement NLP d'un texte
// (tokens, lemmes, entités NER). Nil quand aucun moteur n'est configuré.
type Artifacts struct {
	Tokens      []string
	Lemmas      []string
	NerEntities []NerEntity
}

// Engine est le moteur NLP pluggable (équivalent du NlpEngine du fork).
// L'implémentation ONNX (NER BERT) arrive derrière le build tag `onnx` ;
// NoOp est le défaut 100 % Go pur.
type Engine interface {
	Load() error
	Process(ctx context.Context, text, language string) (*Artifacts, error)
}

// NoOp est un moteur NLP vide : aucun token, aucune entité NER.
type NoOp struct{}

// Load ne fait rien.
func (NoOp) Load() error { return nil }

// Process retourne des artifacts vides.
func (NoOp) Process(_ context.Context, _, _ string) (*Artifacts, error) {
	return &Artifacts{}, nil
}
