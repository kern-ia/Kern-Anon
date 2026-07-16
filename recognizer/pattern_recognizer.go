package recognizer

import (
	"context"
	"errors"
	"unicode/utf8"

	"github.com/YoLaub/presidigo-go/nlp"
	"github.com/YoLaub/presidigo-go/pii"
)

// PatternRecognizer détecte une entité via une liste de patterns regex,
// avec validation optionnelle du match (checksum).
type PatternRecognizer struct {
	name     string
	entity   string
	language string
	patterns []pii.Pattern
	validate ValidateFunc
}

// Option configure un PatternRecognizer.
type Option func(*PatternRecognizer)

// WithValidate installe la fonction de validation des matches.
func WithValidate(f ValidateFunc) Option {
	return func(r *PatternRecognizer) { r.validate = f }
}

// NewPattern construit un PatternRecognizer. Au moins un pattern est requis.
func NewPattern(name, entity, language string, patterns []pii.Pattern, opts ...Option) (*PatternRecognizer, error) {
	if len(patterns) == 0 {
		return nil, errors.New("recognizer: au moins un pattern est requis")
	}
	r := &PatternRecognizer{
		name:     name,
		entity:   entity,
		language: language,
		patterns: patterns,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r, nil
}

func (r *PatternRecognizer) Name() string                { return r.name }
func (r *PatternRecognizer) SupportedEntities() []string { return []string{r.entity} }
func (r *PatternRecognizer) Language() string            { return r.language }

// Analyze applique chaque pattern au texte et retourne les entités détectées,
// avec des offsets exprimés en runes.
func (r *PatternRecognizer) Analyze(_ context.Context, text string, _ *nlp.Artifacts) ([]pii.Result, error) {
	if text == "" {
		return nil, nil
	}
	var results []pii.Result
	for _, p := range r.patterns {
		for _, loc := range p.Regex.FindAllStringIndex(text, -1) {
			score := p.Score
			if r.validate != nil {
				switch v := r.validate(text[loc[0]:loc[1]]); {
				case v == nil:
					// neutre : score du pattern conservé
				case *v:
					score = pii.MaxScore
				default:
					continue
				}
			}
			start := utf8.RuneCountInString(text[:loc[0]])
			end := start + utf8.RuneCountInString(text[loc[0]:loc[1]])
			results = append(results, pii.Result{
				EntityType: r.entity,
				Start:      start,
				End:        end,
				Score:      score,
				Explanation: &pii.Explanation{
					Recognizer:    r.name,
					Pattern:       p.Name,
					OriginalScore: p.Score,
				},
			})
		}
	}
	return results, nil
}
