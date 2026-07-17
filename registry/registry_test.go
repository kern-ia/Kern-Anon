package registry_test

import (
	"regexp"
	"testing"

	"github.com/YoLaub/presidigo-go/pii"
	"github.com/YoLaub/presidigo-go/recognizer"
	"github.com/YoLaub/presidigo-go/registry"
)

func fake(t *testing.T, name, entity, lang string) *recognizer.PatternRecognizer {
	t.Helper()
	r, err := recognizer.NewPattern(name, entity, lang, []pii.Pattern{{
		Name: "p", Regex: regexp.MustCompile(`x`), Score: 0.5,
	}})
	if err != nil {
		t.Fatalf("NewPattern: %v", err)
	}
	return r
}

func TestRegistry_FiltreParLangue(t *testing.T) {
	reg := registry.New()
	reg.Add(fake(t, "EmailEN", "EMAIL_ADDRESS", "en"))
	reg.Add(fake(t, "NirFR", "FR_NIR", "fr"))

	if got := reg.Get("en"); len(got) != 1 || got[0].Name() != "EmailEN" {
		t.Errorf("Get(en) = %v recognizers", len(got))
	}
	if got := reg.Get("fr"); len(got) != 1 || got[0].Name() != "NirFR" {
		t.Errorf("Get(fr) = %v recognizers", len(got))
	}
}

func TestRegistry_FiltreParEntite(t *testing.T) {
	reg := registry.New()
	reg.Add(fake(t, "EmailEN", "EMAIL_ADDRESS", "en"))
	reg.Add(fake(t, "CreditCardEN", "CREDIT_CARD", "en"))

	got := reg.Get("en", "CREDIT_CARD")
	if len(got) != 1 || got[0].Name() != "CreditCardEN" {
		t.Errorf("Get(en, CREDIT_CARD) devrait ne retourner que CreditCardEN, obtenu %d", len(got))
	}
	// Sans filtre d'entité : tous les recognizers de la langue.
	if got := reg.Get("en"); len(got) != 2 {
		t.Errorf("Get(en) = %d, attendu 2", len(got))
	}
}

func TestRegistry_Remove(t *testing.T) {
	reg := registry.New()
	reg.Add(fake(t, "EmailEN", "EMAIL_ADDRESS", "en"))
	reg.Remove("EmailEN")
	if got := reg.Get("en"); len(got) != 0 {
		t.Errorf("après Remove, Get(en) = %d, attendu 0", len(got))
	}
}

func TestDefault_PreRempliParLangue(t *testing.T) {
	reg := registry.Default("en", "fr")
	if got := len(reg.Get("en")); got != 16 { // 7 génériques + 9 US
		t.Errorf("Get(en) = %d recognizers, attendu 16", got)
	}
	if got := len(reg.Get("fr")); got != 12 { // 7 génériques + 5 FR
		t.Errorf("Get(fr) = %d recognizers, attendu 12", got)
	}
	if got := reg.Get("de"); len(got) != 0 {
		t.Errorf("Get(de) = %d, attendu 0 (langue non demandée)", len(got))
	}
}

func TestRegistry_SupportedEntities(t *testing.T) {
	reg := registry.New()
	reg.Add(fake(t, "EmailEN", "EMAIL_ADDRESS", "en"))
	reg.Add(fake(t, "EmailFR", "EMAIL_ADDRESS", "fr"))
	reg.Add(fake(t, "CreditCardEN", "CREDIT_CARD", "en"))

	got := reg.SupportedEntities("en")
	if len(got) != 2 {
		t.Errorf("SupportedEntities(en) = %v, attendu 2 entités distinctes", got)
	}
}
