package recognizer

import (
	"context"
	"errors"
	"sort"

	"github.com/YoLaub/PresidioGo/nlp"
	"github.com/YoLaub/PresidioGo/pii"
)

// PatternRecognizer détecte une entité via une liste de patterns regex,
// avec validation optionnelle du match (checksum).
type PatternRecognizer struct {
	name         string
	entity       string
	language     string
	patterns     []pii.Pattern
	validate     ValidateFunc
	contextWords []string
}

// Option configure un PatternRecognizer.
type Option func(*PatternRecognizer)

// WithValidate installe la fonction de validation des matches.
func WithValidate(f ValidateFunc) Option {
	return func(r *PatternRecognizer) { r.validate = f }
}

// WithContextWords déclare les mots de contexte qui augmentent la confiance
// quand ils apparaissent près de l'entité (exploités par contextaware).
func WithContextWords(words ...string) Option {
	return func(r *PatternRecognizer) { r.contextWords = words }
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

// ContextWords retourne les mots de contexte déclarés (vide si aucun).
func (r *PatternRecognizer) ContextWords() []string { return r.contextWords }

// Analyze applique chaque pattern au texte et retourne les entités détectées,
// avec des offsets exprimés en runes. La conversion bytes→runes se fait en
// une seule passe sur le texte, quel que soit le nombre de matches.
func (r *PatternRecognizer) Analyze(_ context.Context, text string, _ *nlp.Artifacts) ([]pii.Result, error) {
	if text == "" {
		return nil, nil
	}
	type rawMatch struct {
		pattern *pii.Pattern
		score   float64
		b0, b1  int // offsets en bytes
	}
	var raws []rawMatch
	var offsets []int
	for i := range r.patterns {
		p := &r.patterns[i]
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
			raws = append(raws, rawMatch{p, score, loc[0], loc[1]})
			offsets = append(offsets, loc[0], loc[1])
		}
	}
	if len(raws) == 0 {
		return nil, nil
	}

	runeOf := runeOffsets(text, offsets)
	results := make([]pii.Result, 0, len(raws))
	for _, m := range raws {
		results = append(results, pii.Result{
			EntityType: r.entity,
			Start:      runeOf[m.b0],
			End:        runeOf[m.b1],
			Score:      m.score,
			Explanation: &pii.Explanation{
				Recognizer:    r.name,
				Pattern:       m.pattern.Name,
				OriginalScore: m.pattern.Score,
			},
		})
	}
	return results, nil
}

// runeOffsets convertit des offsets en bytes (frontières de runes, ce que
// garantit le moteur regex) en offsets en runes, en une passe sur le texte.
func runeOffsets(text string, offsets []int) map[int]int {
	sorted := make([]int, len(offsets))
	copy(sorted, offsets)
	sort.Ints(sorted)

	out := make(map[int]int, len(sorted))
	j, runeCount := 0, 0
	for byteIdx := range text {
		for j < len(sorted) && sorted[j] <= byteIdx {
			out[sorted[j]] = runeCount
			j++
		}
		if j == len(sorted) {
			return out
		}
		runeCount++
	}
	for j < len(sorted) { // offsets en fin de texte (== len(text))
		out[sorted[j]] = runeCount
		j++
	}
	return out
}
