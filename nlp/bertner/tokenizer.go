// Package bertner porte la partie pure Go du NER BERT : tokenisation
// WordPiece avec offsets en runes, et agrégation des labels BIO en entités.
// L'inférence ONNX elle-même vit dans nlp/onnx (build tag `onnx`).
package bertner

import (
	"bufio"
	"io"
	"strings"
	"unicode"
)

// Token est un token WordPiece : texte (minuscule, préfixe ## pour les
// sous-mots), ID dans le vocab, offsets en runes dans le texte d'origine.
type Token struct {
	Text  string
	ID    int
	Start int
	End   int
}

// Tokenizer est un tokenizer WordPiece basique (équivalent BasicTokenizer +
// WordpieceTokenizer de BERT, sans stripping d'accents).
type Tokenizer struct {
	vocab     map[string]int
	unkID     int
	lowercase bool
}

// NewTokenizer lit un vocab.txt (un token par ligne, ID = numéro de ligne).
func NewTokenizer(vocab io.Reader, lowercase bool) (*Tokenizer, error) {
	m := make(map[string]int)
	sc := bufio.NewScanner(vocab)
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), "\r\n")
		if line == "" {
			continue
		}
		m[line] = len(m)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	unk, ok := m["[UNK]"]
	if !ok {
		unk = 0
	}
	return &Tokenizer{vocab: m, unkID: unk, lowercase: lowercase}, nil
}

// ID retourne l'ID d'un token spécial du vocab ([CLS], [SEP], [PAD]…).
func (t *Tokenizer) ID(token string) int {
	if id, ok := t.vocab[token]; ok {
		return id
	}
	return t.unkID
}

// Tokenize découpe le texte en tokens WordPiece avec offsets en runes.
func (t *Tokenizer) Tokenize(text string) []Token {
	var tokens []Token
	for _, w := range basicSplit(text) {
		tokens = append(tokens, t.wordpiece(w)...)
	}
	return tokens
}

type word struct {
	text  string
	start int // runes
}

// basicSplit découpe sur les espaces et isole la ponctuation, en suivant
// les offsets en runes.
func basicSplit(text string) []word {
	var words []word
	runes := []rune(text)
	i := 0
	for i < len(runes) {
		r := runes[i]
		switch {
		case unicode.IsSpace(r):
			i++
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			words = append(words, word{string(r), i})
			i++
		default:
			start := i
			for i < len(runes) && !unicode.IsSpace(runes[i]) &&
				!unicode.IsPunct(runes[i]) && !unicode.IsSymbol(runes[i]) {
				i++
			}
			words = append(words, word{string(runes[start:i]), start})
		}
	}
	return words
}

// wordpiece découpe un mot en sous-mots du vocab (greedy longest-match),
// ou [UNK] si aucun découpage complet n'existe.
func (t *Tokenizer) wordpiece(w word) []Token {
	text := w.text
	if t.lowercase {
		text = strings.ToLower(text)
	}
	runes := []rune(text)
	wordLen := len([]rune(w.text))

	var pieces []Token
	offset := 0
	for offset < len(runes) {
		end := len(runes)
		var match string
		matchEnd := -1
		for end > offset {
			cand := string(runes[offset:end])
			if offset > 0 {
				cand = "##" + cand
			}
			if _, ok := t.vocab[cand]; ok {
				match, matchEnd = cand, end
				break
			}
			end--
		}
		if matchEnd < 0 {
			// Mot indécomposable : un seul [UNK] couvrant le mot entier.
			return []Token{{Text: "[UNK]", ID: t.unkID, Start: w.start, End: w.start + wordLen}}
		}
		pieces = append(pieces, Token{
			Text:  match,
			ID:    t.vocab[match],
			Start: w.start + offset,
			End:   w.start + matchEnd,
		})
		offset = matchEnd
	}
	return pieces
}

// TokenLabel est le label BIO prédit pour un token, avec sa confiance.
type TokenLabel struct {
	Label string // "O", "B-PER", "I-LOC", …
	Score float64
}

// Entity est une entité NER agrégée, offsets en runes.
type Entity struct {
	Label string // "PER", "LOC", "ORG", "MISC", …
	Start int
	End   int
	Score float64 // moyenne des scores des tokens agrégés
}

// Aggregate fusionne les labels BIO token par token en entités : B-X ouvre,
// I-X prolonge (ou ouvre si orphelin), O ferme. Le score est la moyenne.
func Aggregate(tokens []Token, labels []TokenLabel) []Entity {
	var entities []Entity
	var cur *Entity
	var scores []float64

	flush := func() {
		if cur != nil {
			sum := 0.0
			for _, s := range scores {
				sum += s
			}
			cur.Score = sum / float64(len(scores))
			entities = append(entities, *cur)
			cur, scores = nil, nil
		}
	}

	for i, tok := range tokens {
		if i >= len(labels) {
			break
		}
		label := labels[i]
		// Sous-mot de continuation d'une entité ouverte : le mot reste
		// entier même si le modèle étiquette O ce fragment.
		if cur != nil && strings.HasPrefix(tok.Text, "##") && tok.Start == cur.End {
			cur.End = tok.End
			if strings.HasPrefix(label.Label, "I-") && label.Label[2:] == cur.Label {
				scores = append(scores, label.Score)
			}
			continue
		}
		switch {
		case label.Label == "O" || label.Label == "":
			flush()
		case strings.HasPrefix(label.Label, "B-"):
			flush()
			cur = &Entity{Label: label.Label[2:], Start: tok.Start, End: tok.End}
			scores = []float64{label.Score}
		case strings.HasPrefix(label.Label, "I-"):
			kind := label.Label[2:]
			if cur != nil && cur.Label == kind {
				cur.End = tok.End
				scores = append(scores, label.Score)
			} else {
				flush()
				cur = &Entity{Label: kind, Start: tok.Start, End: tok.End}
				scores = []float64{label.Score}
			}
		default:
			flush()
		}
	}
	flush()
	return entities
}
