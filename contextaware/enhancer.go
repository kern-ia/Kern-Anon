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
// suffixCount mots après (offsets en runes). Le balayage est borné au
// voisinage de l'entité — jamais tout le préfixe du texte.
func (e *Enhancer) window(runes []rune, start, end int) string {
	before := wordsBackward(runes, min(max(start, 0), len(runes)), e.prefixCount)
	after := wordsForward(runes, min(max(end, 0), len(runes)), e.suffixCount)
	return strings.ToLower(strings.Join(append(before, after...), " "))
}

func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

// wordsBackward collecte jusqu'à n mots en remontant depuis from (exclu),
// dans l'ordre du texte.
func wordsBackward(runes []rune, from, n int) []string {
	if n <= 0 {
		return nil
	}
	words := make([]string, 0, n)
	i := from - 1
	for i >= 0 && len(words) < n {
		for i >= 0 && !isWordRune(runes[i]) {
			i--
		}
		if i < 0 {
			break
		}
		end := i + 1
		for i >= 0 && isWordRune(runes[i]) {
			i--
		}
		words = append(words, string(runes[i+1:end]))
	}
	for l, r := 0, len(words)-1; l < r; l, r = l+1, r-1 {
		words[l], words[r] = words[r], words[l]
	}
	return words
}

// wordsForward collecte jusqu'à n mots à partir de from (inclus).
func wordsForward(runes []rune, from, n int) []string {
	if n <= 0 {
		return nil
	}
	words := make([]string, 0, n)
	i := from
	for i < len(runes) && len(words) < n {
		for i < len(runes) && !isWordRune(runes[i]) {
			i++
		}
		start := i
		for i < len(runes) && isWordRune(runes[i]) {
			i++
		}
		if i > start {
			words = append(words, string(runes[start:i]))
		}
	}
	return words
}

func containsAny(window string, words []string) bool {
	for _, w := range words {
		if w != "" && strings.Contains(window, strings.ToLower(w)) {
			return true
		}
	}
	return false
}
