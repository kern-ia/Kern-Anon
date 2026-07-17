package anonymizer_test

import (
	"strings"
	"testing"

	"github.com/YoLaub/presidigo-go/anonymizer"
	"github.com/YoLaub/presidigo-go/pii"
)

func res(entity string, start, end int, score float64) pii.Result {
	return pii.Result{EntityType: entity, Start: start, End: end, Score: score}
}

func TestAnonymize_ReplaceParDefaut(t *testing.T) {
	// Sans opérateur configuré : replace par <ENTITY_TYPE> (comportement Presidio).
	eng := anonymizer.New()
	out, err := eng.Anonymize("call 0601020304 now",
		[]pii.Result{res("PHONE_NUMBER", 5, 15, 0.7)}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out.Text != "call <PHONE_NUMBER> now" {
		t.Errorf("Text = %q", out.Text)
	}
	if len(out.Items) != 1 || out.Items[0].Operator != "replace" {
		t.Errorf("Items = %+v", out.Items)
	}
}

func TestAnonymize_OperateurParEntiteEtDefault(t *testing.T) {
	eng := anonymizer.New()
	out, err := eng.Anonymize("mail info@presidio.site tel 0601020304",
		[]pii.Result{
			res("EMAIL_ADDRESS", 5, 23, 0.5),
			res("PHONE_NUMBER", 28, 38, 0.7),
		},
		map[string]anonymizer.Operator{
			"EMAIL_ADDRESS": anonymizer.Replace("<MAIL>"),
			"DEFAULT":       anonymizer.Redact(),
		})
	if err != nil {
		t.Fatal(err)
	}
	if out.Text != "mail <MAIL> tel " {
		t.Errorf("Text = %q", out.Text)
	}
}

func TestAnonymize_OffsetsRunes(t *testing.T) {
	// Les offsets d'entrée sont en runes ; les accents ne décalent rien.
	eng := anonymizer.New()
	text := "Prénom : José — email : info@presidio.site"
	out, err := eng.Anonymize(text,
		[]pii.Result{res("EMAIL_ADDRESS", 24, 42, 0.5)},
		map[string]anonymizer.Operator{"DEFAULT": anonymizer.Replace("<MAIL>")})
	if err != nil {
		t.Fatal(err)
	}
	if out.Text != "Prénom : José — email : <MAIL>" {
		t.Errorf("Text = %q", out.Text)
	}
	// Les items exposent les NOUVEAUX offsets (en runes) dans le texte de sortie.
	if len(out.Items) != 1 || out.Items[0].Start != 24 || out.Items[0].End != 30 {
		t.Errorf("Items = %+v", out.Items)
	}
}

func TestAnonymize_ChevauchementScoreLePlusHautGagne(t *testing.T) {
	eng := anonymizer.New()
	// URL contenue dans l'email : l'email (score plus haut) doit gagner,
	// une seule substitution.
	out, err := eng.Anonymize("contact: info@presidio.site",
		[]pii.Result{
			res("EMAIL_ADDRESS", 9, 27, 0.7),
			res("URL", 14, 27, 0.5),
		},
		map[string]anonymizer.Operator{"DEFAULT": anonymizer.Replace("<X>")})
	if err != nil {
		t.Fatal(err)
	}
	if out.Text != "contact: <X>" {
		t.Errorf("Text = %q", out.Text)
	}
	if len(out.Items) != 1 || out.Items[0].EntityType != "EMAIL_ADDRESS" {
		t.Errorf("Items = %+v", out.Items)
	}
}

func TestAnonymize_ChevauchementMemeScoreLePlusLongGagne(t *testing.T) {
	eng := anonymizer.New()
	out, err := eng.Anonymize("contact: info@presidio.site",
		[]pii.Result{
			res("URL", 14, 27, 0.5),
			res("EMAIL_ADDRESS", 9, 27, 0.5),
		},
		map[string]anonymizer.Operator{"DEFAULT": anonymizer.Replace("<X>")})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Items) != 1 || out.Items[0].EntityType != "EMAIL_ADDRESS" {
		t.Errorf("le span le plus long doit gagner à score égal : %+v", out.Items)
	}
}

func TestMask(t *testing.T) {
	eng := anonymizer.New()
	out, err := eng.Anonymize("card 4012888888881881",
		[]pii.Result{res("CREDIT_CARD", 5, 21, 1.0)},
		map[string]anonymizer.Operator{"DEFAULT": anonymizer.Mask('*', 12, false)})
	if err != nil {
		t.Fatal(err)
	}
	if out.Text != "card ************1881" {
		t.Errorf("Text = %q", out.Text)
	}
}

func TestMask_FromEnd(t *testing.T) {
	eng := anonymizer.New()
	out, err := eng.Anonymize("card 4012888888881881",
		[]pii.Result{res("CREDIT_CARD", 5, 21, 1.0)},
		map[string]anonymizer.Operator{"DEFAULT": anonymizer.Mask('*', 4, true)})
	if err != nil {
		t.Fatal(err)
	}
	if out.Text != "card 401288888888****" {
		t.Errorf("Text = %q", out.Text)
	}
}

func TestHash_Sha256Deterministe(t *testing.T) {
	eng := anonymizer.New()
	ops := map[string]anonymizer.Operator{"DEFAULT": anonymizer.Hash()}
	out1, err := eng.Anonymize("x info@presidio.site",
		[]pii.Result{res("EMAIL_ADDRESS", 2, 20, 0.5)}, ops)
	if err != nil {
		t.Fatal(err)
	}
	out2, _ := eng.Anonymize("x info@presidio.site",
		[]pii.Result{res("EMAIL_ADDRESS", 2, 20, 0.5)}, ops)
	if out1.Text != out2.Text {
		t.Error("le hash doit être déterministe")
	}
	if strings.Contains(out1.Text, "info@") {
		t.Errorf("PII encore présente : %q", out1.Text)
	}
	if len(out1.Text) != len("x ")+64 { // sha256 hex
		t.Errorf("longueur inattendue : %q", out1.Text)
	}
}

func TestKeep(t *testing.T) {
	eng := anonymizer.New()
	out, err := eng.Anonymize("id 123",
		[]pii.Result{res("ID", 3, 6, 0.9)},
		map[string]anonymizer.Operator{"DEFAULT": anonymizer.Keep()})
	if err != nil {
		t.Fatal(err)
	}
	if out.Text != "id 123" {
		t.Errorf("Keep ne doit rien changer : %q", out.Text)
	}
	if len(out.Items) != 1 {
		t.Errorf("Keep doit quand même tracer l'item : %+v", out.Items)
	}
}

func TestCustom(t *testing.T) {
	eng := anonymizer.New()
	out, err := eng.Anonymize("bonjour José",
		[]pii.Result{res("PERSON", 8, 12, 0.85)},
		map[string]anonymizer.Operator{
			"DEFAULT": anonymizer.Custom("upper", func(v string) (string, error) {
				return strings.ToUpper(v), nil
			}),
		})
	if err != nil {
		t.Fatal(err)
	}
	if out.Text != "bonjour JOSÉ" {
		t.Errorf("Text = %q", out.Text)
	}
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef") // 32 octets → AES-256-GCM
	eng := anonymizer.New()

	enc, err := anonymizer.Encrypt(key)
	if err != nil {
		t.Fatal(err)
	}
	out, err := eng.Anonymize("mail info@presidio.site",
		[]pii.Result{res("EMAIL_ADDRESS", 5, 23, 0.5)},
		map[string]anonymizer.Operator{"DEFAULT": enc})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.Text, "info@") {
		t.Fatalf("PII encore présente : %q", out.Text)
	}

	dec, err := anonymizer.Decrypt(key)
	if err != nil {
		t.Fatal(err)
	}
	back, err := eng.Deanonymize(out.Text, out.Items,
		map[string]anonymizer.Operator{"DEFAULT": dec})
	if err != nil {
		t.Fatal(err)
	}
	if back.Text != "mail info@presidio.site" {
		t.Errorf("round-trip raté : %q", back.Text)
	}
}

func TestEncrypt_CleInvalide(t *testing.T) {
	if _, err := anonymizer.Encrypt([]byte("trop-courte")); err == nil {
		t.Error("clé de longueur invalide : erreur attendue")
	}
}
