// Package nlp définit les artefacts issus du traitement NLP d'un texte,
// partagés entre tous les recognizers. L'interface du moteur (NlpEngine)
// et son implémentation ONNX arrivent avec la feature nlp-onnx.
package nlp

// Artifacts porte les résultats du traitement NLP d'un texte
// (tokens, lemmes, entités NER). Nil quand aucun moteur n'est configuré.
type Artifacts struct {
	Tokens []string
	Lemmas []string
}
