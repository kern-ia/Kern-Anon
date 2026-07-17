package bertner_test

import (
	"strings"
	"testing"

	"github.com/YoLaub/presidigo-go/nlp/bertner"
)

// Petit vocabulaire de test : suffisant pour vérifier la mécanique WordPiece
// sans embarquer le vocab BERT complet.
const testVocab = `[PAD]
[UNK]
[CLS]
[SEP]
john
works
at
acme
in
paris
##son
mary
и`

func newTok(t *testing.T) *bertner.Tokenizer {
	t.Helper()
	tok, err := bertner.NewTokenizer(strings.NewReader(testVocab), true)
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func TestTokenize_MotsSimples(t *testing.T) {
	tok := newTok(t)
	tokens := tok.Tokenize("John works at ACME")
	want := []string{"john", "works", "at", "acme"}
	if len(tokens) != len(want) {
		t.Fatalf("tokens = %+v", tokens)
	}
	for i, w := range want {
		if tokens[i].Text != w {
			t.Errorf("token %d = %q, attendu %q", i, tokens[i].Text, w)
		}
	}
	// IDs : d'après l'ordre du vocab (0-indexé).
	if tokens[0].ID != 4 || tokens[3].ID != 7 {
		t.Errorf("IDs inattendus : %+v", tokens)
	}
}

func TestTokenize_SousMotsWordPiece(t *testing.T) {
	tok := newTok(t)
	tokens := tok.Tokenize("Johnson")
	if len(tokens) != 2 || tokens[0].Text != "john" || tokens[1].Text != "##son" {
		t.Fatalf("attendu [john ##son], obtenu %+v", tokens)
	}
	// Les deux sous-mots couvrent le même mot d'origine : offsets contigus en runes.
	if tokens[0].Start != 0 || tokens[0].End != 4 || tokens[1].Start != 4 || tokens[1].End != 7 {
		t.Errorf("offsets = %+v", tokens)
	}
}

func TestTokenize_OffsetsRunesAvecAccents(t *testing.T) {
	tok := newTok(t)
	// « é » (2 bytes) avant le mot : les offsets doivent être en runes.
	tokens := tok.Tokenize("éé john")
	var john *bertner.Token
	for i := range tokens {
		if tokens[i].Text == "john" {
			john = &tokens[i]
		}
	}
	if john == nil {
		t.Fatalf("john non trouvé : %+v", tokens)
	}
	if john.Start != 3 || john.End != 7 {
		t.Errorf("offsets runes = (%d,%d), attendu (3,7)", john.Start, john.End)
	}
}

func TestTokenize_InconnuDonneUNK(t *testing.T) {
	tok := newTok(t)
	tokens := tok.Tokenize("zzzzz")
	if len(tokens) != 1 || tokens[0].ID != 1 { // [UNK]
		t.Fatalf("attendu [UNK], obtenu %+v", tokens)
	}
}

func TestTokenize_PonctuationSeparee(t *testing.T) {
	tok := newTok(t)
	tokens := tok.Tokenize("john, works.")
	// La ponctuation devient des tokens séparés ([UNK] avec ce vocab).
	if len(tokens) != 4 {
		t.Fatalf("attendu 4 tokens (john , works .), obtenu %+v", tokens)
	}
	if tokens[0].Text != "john" || tokens[2].Text != "works" {
		t.Errorf("tokens = %+v", tokens)
	}
}

func TestAggregate_BIOVersEntites(t *testing.T) {
	// Simule la sortie du modèle : B-PER sur "john", I-PER sur "##son",
	// O sur le reste → une seule entité PER couvrant "Johnson".
	tok := newTok(t)
	tokens := tok.Tokenize("Johnson works in Paris")
	labels := make([]bertner.TokenLabel, len(tokens))
	for i := range labels {
		labels[i] = bertner.TokenLabel{Label: "O", Score: 0.99}
	}
	labels[0] = bertner.TokenLabel{Label: "B-PER", Score: 0.98}
	labels[1] = bertner.TokenLabel{Label: "I-PER", Score: 0.96}
	// "paris" est le dernier token.
	labels[len(labels)-1] = bertner.TokenLabel{Label: "B-LOC", Score: 0.90}

	entities := bertner.Aggregate(tokens, labels)
	if len(entities) != 2 {
		t.Fatalf("attendu 2 entités, obtenu %+v", entities)
	}
	per := entities[0]
	if per.Label != "PER" || per.Start != 0 || per.End != 7 {
		t.Errorf("PER = %+v", per)
	}
	if per.Score < 0.95 || per.Score > 0.99 {
		t.Errorf("score PER = moyenne attendue ≈0.97, obtenu %v", per.Score)
	}
	loc := entities[1]
	if loc.Label != "LOC" || loc.Start != 17 || loc.End != 22 {
		t.Errorf("LOC = %+v", loc)
	}
}

func TestAggregate_EtendJusquAuBoutDuMot(t *testing.T) {
	// Le modèle étiquette parfois O le dernier sous-mot d'un nom : l'entité
	// doit quand même couvrir le mot entier (jamais de coupure intra-mot).
	tok := newTok(t)
	tokens := tok.Tokenize("Johnson")
	if len(tokens) != 2 {
		t.Fatalf("attendu [john ##son], obtenu %+v", tokens)
	}
	entities := bertner.Aggregate(tokens, []bertner.TokenLabel{
		{Label: "B-PER", Score: 0.95},
		{Label: "O", Score: 0.99}, // ##son étiqueté O par le modèle
	})
	if len(entities) != 1 || entities[0].End != 7 {
		t.Fatalf("l'entité doit couvrir tout « Johnson » [0:7], obtenu %+v", entities)
	}
}

func TestAggregate_IOrphelinDemarreEntite(t *testing.T) {
	// Un I-X sans B-X précédent démarre quand même une entité (robustesse).
	tok := newTok(t)
	tokens := tok.Tokenize("paris")
	entities := bertner.Aggregate(tokens, []bertner.TokenLabel{{Label: "I-LOC", Score: 0.9}})
	if len(entities) != 1 || entities[0].Label != "LOC" {
		t.Fatalf("entités = %+v", entities)
	}
}
