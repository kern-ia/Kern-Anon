package recognizer_test

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/YoLaub/PresidioGo/pii"
	"github.com/YoLaub/PresidioGo/recognizer"
)

// benchText : ~60 Ko de texte accentué (multi-bytes) truffé d'emails — le cas
// pathologique de la conversion bytes→runes par match.
func benchText() string {
	block := "Prénom : José, préférence élevée — contact : agent"
	var b strings.Builder
	for i := 0; i < 400; i++ {
		b.WriteString(block)
		b.WriteString(" info@presidio.site ; ")
		b.WriteString(block)
		b.WriteString(" autre.adresse@exemple.org ! ")
	}
	return b.String()
}

func BenchmarkPatternRecognizer_TexteLongMultiMatch(b *testing.B) {
	r, err := recognizer.NewPattern("EmailRecognizer", "EMAIL_ADDRESS", "en",
		[]pii.Pattern{{
			Name:  "email-basic",
			Regex: regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`),
			Score: 0.6,
		}})
	if err != nil {
		b.Fatal(err)
	}
	text := benchText()
	b.SetBytes(int64(len(text)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results, err := r.Analyze(context.Background(), text, nil)
		if err != nil {
			b.Fatal(err)
		}
		if len(results) != 800 {
			b.Fatalf("attendu 800 matches, obtenu %d", len(results))
		}
	}
}
