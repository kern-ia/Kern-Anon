package analyzer_test

import (
	"context"
	"strings"
	"testing"

	"github.com/YoLaub/PresidioGo/analyzer"
	"github.com/YoLaub/PresidioGo/registry"
)

// benchDoc : document réaliste (~30 Ko) mêlant PII et texte accentué,
// analysé par le registry complet en+fr (28 recognizers).
func benchDoc() string {
	para := "Le client José Martin (email : jose.martin@exemple.org, tél 06 01 02 03 04) " +
		"a réglé avec la carte 4012-8888-8888-1881 depuis l'IP 192.168.0.1. " +
		"IBAN GB29NWBK60161331926819, siren 552100554, ssn 216-09-1234, " +
		"url https://www.microsoft.com/fr — dossier archivé. "
	var b strings.Builder
	for i := 0; i < 100; i++ {
		b.WriteString(para)
	}
	return b.String()
}

func BenchmarkAnalyze_RegistryComplet(b *testing.B) {
	eng, err := analyzer.New(analyzer.WithRegistry(registry.Default("en", "fr")))
	if err != nil {
		b.Fatal(err)
	}
	text := benchDoc()
	ctx := context.Background()
	b.SetBytes(int64(len(text)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results, err := eng.Analyze(ctx, text, analyzer.Language("en"))
		if err != nil {
			b.Fatal(err)
		}
		if len(results) == 0 {
			b.Fatal("aucun résultat")
		}
	}
}
