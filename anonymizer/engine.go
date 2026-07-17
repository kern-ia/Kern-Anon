package anonymizer

import (
	"fmt"
	"sort"

	"github.com/YoLaub/PresidioGo/pii"
)

// Item décrit une substitution effectuée : offsets (en runes) DANS LE TEXTE
// DE SORTIE, entité d'origine et opérateur appliqué.
type Item struct {
	EntityType string
	Start      int
	End        int
	Operator   string
}

// Result est le produit d'une anonymisation.
type Result struct {
	Text  string
	Items []Item
}

// Engine applique les opérateurs aux entités détectées.
type Engine struct{}

// New crée un moteur d'anonymisation.
func New() *Engine {
	return &Engine{}
}

// Anonymize remplace chaque entité de results dans text par la sortie de son
// opérateur. L'opérateur est choisi par type d'entité, sinon "DEFAULT",
// sinon Replace("<ENTITY_TYPE>"). Les chevauchements sont résolus avant
// substitution : score le plus haut, puis span le plus long.
func (e *Engine) Anonymize(text string, results []pii.Result, operators map[string]Operator) (*Result, error) {
	kept := resolveConflicts(results)

	runes := []rune(text)
	out := make([]rune, 0, len(runes))
	var items []Item
	cursor := 0

	for _, r := range kept {
		if r.Start < 0 || r.End > len(runes) || r.Start > r.End {
			return nil, fmt.Errorf("anonymizer: offsets hors bornes [%d:%d] pour un texte de %d runes",
				r.Start, r.End, len(runes))
		}
		op := pickOperator(operators, r.EntityType)
		replaced, err := op.Operate(string(runes[r.Start:r.End]))
		if err != nil {
			return nil, fmt.Errorf("anonymizer: opérateur %s sur %s : %w", op.Name(), r.EntityType, err)
		}
		out = append(out, runes[cursor:r.Start]...)
		newStart := len(out)
		out = append(out, []rune(replaced)...)
		items = append(items, Item{
			EntityType: r.EntityType,
			Start:      newStart,
			End:        len(out),
			Operator:   op.Name(),
		})
		cursor = r.End
	}
	out = append(out, runes[cursor:]...)

	return &Result{Text: string(out), Items: items}, nil
}

// Deanonymize applique des opérateurs (typiquement Decrypt) aux items issus
// d'une anonymisation précédente, offsets exprimés dans text.
func (e *Engine) Deanonymize(text string, items []Item, operators map[string]Operator) (*Result, error) {
	results := make([]pii.Result, len(items))
	for i, it := range items {
		results[i] = pii.Result{
			EntityType: it.EntityType,
			Start:      it.Start,
			End:        it.End,
			Score:      pii.MaxScore,
		}
	}
	return e.Anonymize(text, results, operators)
}

func pickOperator(operators map[string]Operator, entityType string) Operator {
	if op, ok := operators[entityType]; ok {
		return op
	}
	if op, ok := operators["DEFAULT"]; ok {
		return op
	}
	return Replace("<" + entityType + ">")
}

// resolveConflicts garde, parmi les résultats qui se chevauchent, celui au
// score le plus haut (puis le plus long, puis le plus à gauche), et retourne
// les résultats retenus triés par position.
func resolveConflicts(results []pii.Result) []pii.Result {
	byPriority := make([]pii.Result, len(results))
	copy(byPriority, results)
	sort.SliceStable(byPriority, func(i, j int) bool {
		a, b := byPriority[i], byPriority[j]
		if a.Score != b.Score {
			return a.Score > b.Score
		}
		if la, lb := a.End-a.Start, b.End-b.Start; la != lb {
			return la > lb
		}
		return a.Start < b.Start
	})

	var kept []pii.Result
	for _, cand := range byPriority {
		conflict := false
		for _, k := range kept {
			if cand.Start < k.End && k.Start < cand.End {
				conflict = true
				break
			}
		}
		if !conflict {
			kept = append(kept, cand)
		}
	}
	sort.Slice(kept, func(i, j int) bool { return kept[i].Start < kept[j].Start })
	return kept
}
