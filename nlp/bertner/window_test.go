package bertner_test

import (
	"testing"

	"github.com/YoLaub/presidigo-go/nlp/bertner"
)

func mkTokens(n int) []bertner.Token {
	tokens := make([]bertner.Token, n)
	for i := range tokens {
		tokens[i] = bertner.Token{Text: "x", ID: 1, Start: i * 2, End: i*2 + 1}
	}
	return tokens
}

func TestWindows_CourtUneSeuleFenetre(t *testing.T) {
	tokens := mkTokens(100)
	wins := bertner.Windows(tokens, 510, 382)
	if len(wins) != 1 || len(wins[0]) != 100 {
		t.Fatalf("attendu 1 fenêtre de 100, obtenu %d", len(wins))
	}
}

func TestWindows_LongAvecChevauchement(t *testing.T) {
	tokens := mkTokens(1000)
	wins := bertner.Windows(tokens, 510, 382)
	if len(wins) != 3 {
		t.Fatalf("attendu 3 fenêtres (0, 382, 764), obtenu %d", len(wins))
	}
	if len(wins[0]) != 510 || len(wins[1]) != 510 {
		t.Errorf("tailles = %d, %d", len(wins[0]), len(wins[1]))
	}
	// Dernière fenêtre : couvre la fin (764 → 1000).
	last := wins[2]
	if last[len(last)-1].Start != tokens[999].Start {
		t.Errorf("la dernière fenêtre doit atteindre le dernier token")
	}
	// Chevauchement : la fenêtre 2 recouvre les 128 derniers tokens de la 1.
	if wins[1][0].Start != tokens[382].Start {
		t.Errorf("stride attendu 382, fenêtre 2 démarre à %d", wins[1][0].Start)
	}
}

func TestWindows_StrideInvalideCorrige(t *testing.T) {
	// stride >= size ou <= 0 ne doit ni boucler ni paniquer.
	tokens := mkTokens(20)
	if got := bertner.Windows(tokens, 10, 0); len(got) < 2 {
		t.Errorf("stride 0 corrigé attendu, fenêtres = %d", len(got))
	}
}

func TestMergeEntities_DoublonsDesFenetres(t *testing.T) {
	entities := []bertner.Entity{
		{Label: "PER", Start: 10, End: 20, Score: 0.90}, // fenêtre 1
		{Label: "PER", Start: 10, End: 20, Score: 0.95}, // même entité, fenêtre 2
		{Label: "LOC", Start: 40, End: 45, Score: 0.80},
	}
	merged := bertner.MergeEntities(entities)
	if len(merged) != 2 {
		t.Fatalf("attendu 2 entités après fusion, obtenu %+v", merged)
	}
	if merged[0].Label != "PER" || merged[0].Score != 0.95 {
		t.Errorf("le doublon au meilleur score doit gagner : %+v", merged[0])
	}
}

func TestMergeEntities_ChevauchementPartielMemeLabel(t *testing.T) {
	// Entité tronquée en bord de fenêtre 1, complète en fenêtre 2.
	entities := []bertner.Entity{
		{Label: "PER", Start: 10, End: 15, Score: 0.90}, // tronquée
		{Label: "PER", Start: 10, End: 22, Score: 0.85}, // complète
	}
	merged := bertner.MergeEntities(entities)
	if len(merged) != 1 || merged[0].End != 22 {
		t.Fatalf("le span le plus long doit gagner : %+v", merged)
	}
}

func TestMergeEntities_LabelsDifferentsConserves(t *testing.T) {
	entities := []bertner.Entity{
		{Label: "PER", Start: 10, End: 20, Score: 0.9},
		{Label: "ORG", Start: 15, End: 25, Score: 0.9},
	}
	if merged := bertner.MergeEntities(entities); len(merged) != 2 {
		t.Fatalf("labels différents : pas de fusion, obtenu %+v", merged)
	}
}
