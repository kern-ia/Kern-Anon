package pii

import "regexp"

// MaxScore est le score attribué à un résultat confirmé par validation
// (ex. checksum Luhn valide).
const MaxScore = 1.0

// Pattern associe une expression régulière à un score de confiance de base.
type Pattern struct {
	Name  string
	Regex *regexp.Regexp
	Score float64
}

// Explanation trace la provenance d'un résultat : quel recognizer, quel
// pattern, et les ajustements de score appliqués.
type Explanation struct {
	Recognizer    string
	Pattern       string
	OriginalScore float64
	ContextBoost  float64
}

// Result est une entité PII détectée. Start et End sont des offsets en runes.
type Result struct {
	EntityType  string
	Start       int
	End         int
	Score       float64
	Explanation *Explanation
}
