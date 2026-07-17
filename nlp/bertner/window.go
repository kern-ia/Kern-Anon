package bertner

import "sort"

// Windows découpe les tokens en fenêtres de taille size avec un pas stride
// (stride < size ⇒ chevauchement). La dernière fenêtre est recalée pour
// couvrir la fin. Un stride invalide est ramené à size/2.
func Windows(tokens []Token, size, stride int) [][]Token {
	if len(tokens) <= size {
		return [][]Token{tokens}
	}
	if stride <= 0 || stride >= size {
		stride = max(size/2, 1)
	}
	var wins [][]Token
	for start := 0; ; start += stride {
		end := start + size
		if end >= len(tokens) {
			wins = append(wins, tokens[len(tokens)-size:])
			return wins
		}
		wins = append(wins, tokens[start:end])
	}
}

// MergeEntities fusionne les entités issues de fenêtres chevauchantes :
// deux entités de même label qui se chevauchent sont réduites à la
// meilleure (span le plus long, puis score le plus haut).
func MergeEntities(entities []Entity) []Entity {
	if len(entities) <= 1 {
		return entities
	}
	byPriority := make([]Entity, len(entities))
	copy(byPriority, entities)
	sort.SliceStable(byPriority, func(i, j int) bool {
		a, b := byPriority[i], byPriority[j]
		if la, lb := a.End-a.Start, b.End-b.Start; la != lb {
			return la > lb
		}
		return a.Score > b.Score
	})

	var kept []Entity
	for _, cand := range byPriority {
		overlaps := false
		for _, k := range kept {
			if k.Label == cand.Label && cand.Start < k.End && k.Start < cand.End {
				overlaps = true
				break
			}
		}
		if !overlaps {
			kept = append(kept, cand)
		}
	}
	sort.Slice(kept, func(i, j int) bool { return kept[i].Start < kept[j].Start })
	return kept
}
