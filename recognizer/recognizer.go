// Package recognizer définit l'interface Recognizer et l'implémentation
// générique PatternRecognizer (regex + validation par checksum).
package recognizer

import (
	"context"

	"github.com/YoLaub/presidigo-go/nlp"
	"github.com/YoLaub/presidigo-go/pii"
)

// Recognizer détecte des entités PII dans un texte.
type Recognizer interface {
	Name() string
	SupportedEntities() []string
	Language() string
	Analyze(ctx context.Context, text string, artifacts *nlp.Artifacts) ([]pii.Result, error)
}

// ValidateFunc valide un match : nil = neutre (score du pattern conservé),
// true = confirmé (score porté à pii.MaxScore), false = rejeté.
type ValidateFunc func(match string) *bool
