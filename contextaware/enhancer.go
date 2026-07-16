// Package contextaware augmente le score des entités détectées quand des
// mots de contexte du recognizer apparaissent près de l'entité — portage du
// LemmaContextAwareEnhancer du fork (mode « substring », sans lemmatisation
// tant que le moteur NLP n'est pas branché).
package contextaware

import (
	"strings"
	"unicode"

	"github.com/YoLaub/presidigo-go/pii"
	"github.com/YoLaub/presidigo-go/recognizer"
)

// ContextAware est implémenté par les recognizers qui déclarent des mots de
// contexte (PatternRecognizer via WithContextWords).
type ContextAware interface {
	ContextWords() []string
}

// Enhancer applique le boost contextuel. Valeurs par défaut du fork :
// facteur 0.35, score minimum avec contexte 0.4, fenêtre de 5 mots avant
// l'entité et 0 après.
type Enhancer struct {
	similarityFactor    float64
	minScoreWithContext float64
	prefixCount         int
	suffixCount         int
}

// Option configure l'Enhancer.
type Option func(*Enhancer)

// WithSimilarityFactor change le boost appliqué (défaut 0.35).
func WithSimilarityFactor(f float64) Option {
	return func(e *Enhancer) { e.similarityFactor = f }
}

// WithWindow change la fenêtre de recherche en mots (défaut 5 avant, 0 après).
func WithWindow(before, after int) Option {
	return func(e *Enhancer) { e.prefixCount = before; e.suffixCount = after }
}

// New crée un Enhancer avec les valeurs par défaut du fork.
func New(opts ...Option) *Enhancer {
	e := &Enhancer{
		similarityFactor:    0.35,
		minScoreWithContext: 0.4,
		prefixCount:         5,
		suffixCount:         0,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Enhance retourne les résultats avec score éventuellement augmenté. Le
// recognizer d'origine est retrouvé via Explanation.Recognizer ; sans mots
// de contexte déclarés, le résultat est inchangé.
func (e *Enhancer) Enhance(text string, results []pii.Result, recognizers []recognizer.Recognizer) []pii.Result {
	words := make(map[string][]string, len(recognizers))
	for _, rec := range recognizers {
		if ca, ok := rec.(ContextAware); ok {
			words[rec.Name()] = ca.ContextWords()
		}
	}

	runes := []rune(text)
	out := make([]pii.Result, len(results))
	for i, r := range results {
		out[i] = r
		if r.Explanation == nil {
			continue
		}
		ctxWords := words[r.Explanation.Recognizer]
		if len(ctxWords) == 0 {
			continue
		}
		window := e.window(runes, r.Start, r.End)
		if !containsAny(window, ctxWords) {
			continue
		}
		boosted := r.Score + e.similarityFactor
		boosted = max(boosted, e.minScoreWithContext)
		boosted = min(boosted, pii.MaxScore)
		out[i].Score = boosted
		expl := *r.Explanation
		expl.ContextBoost = e.similarityFactor
		out[i].Explanation = &expl
	}
	return out
}

// window retourne, en minuscules, les prefixCount mots avant l'entité et les
// suffixCount mots après (offsets en runes).
func (e *Enhancer) window(runes []rune, start, end int) string {
	before := lastWords(string(runes[:max(start, 0)]), e.prefixCount)
	after := firstWords(string(runes[min(end, len(runes)):]), e.suffixCount)
	return strings.ToLower(strings.Join(append(before, after...), " "))
}

func splitWords(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

func lastWords(s string, n int) []string {
	w := splitWords(s)
	if len(w) > n {
		w = w[len(w)-n:]
	}
	return w
}

func firstWords(s string, n int) []string {
	w := splitWords(s)
	if len(w) > n {
		w = w[:n]
	}
	return w
}

func containsAny(window string, words []string) bool {
	for _, w := range words {
		if w != "" && strings.Contains(window, strings.ToLower(w)) {
			return true
		}
	}
	return false
}
