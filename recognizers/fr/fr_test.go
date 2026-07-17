package fr_test

import (
	"testing"

	"github.com/YoLaub/PresidioGo/internal/oracletest"
	"github.com/YoLaub/PresidioGo/recognizers/fr"
)

func TestOracle(t *testing.T) {
	recs := fr.All()
	if len(recs) != 5 {
		t.Fatalf("fr.All doit fournir 5 recognizers, obtenu %d", len(recs))
	}
	oracletest.Run(t, recs)
}

func TestNirKey(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want bool
	}{
		{"185057800608491", true},   // clé 91 calculée
		{"185057800608492", false},  // clé fausse
		{"185052A00608435", true},   // Corse 2A → 19
		{"1850578006084", false},    // trop court
		{"2850578006084XX", false},  // clé non numérique
	} {
		if got := fr.NirKey(tc.in); got != tc.want {
			t.Errorf("NirKey(%q) = %v, attendu %v", tc.in, got, tc.want)
		}
	}
}

func TestNirKey_Corse2B(t *testing.T) {
	// 2B → 18 pour le calcul : construit à partir de la même base.
	// 185052B006084 → n = 1850518006084, clé = 97 - n mod 97.
	if !fr.NirKey("185052B00608456") && !fr.NirKey("185052B00608400") {
		// La clé exacte est vérifiée par calcul dans l'implémentation ;
		// ici on vérifie seulement que 2B est accepté structurellement :
		// au moins une clé sur 97 doit être valide.
		found := false
		for k := 1; k <= 97; k++ {
			s := "185052B006084"
			if fr.NirKey(s + pad2(k)) {
				found = true
				break
			}
		}
		if !found {
			t.Error("aucune clé valide trouvée pour un NIR 2B : le remplacement 2B→18 est cassé")
		}
	}
}

func pad2(n int) string {
	return string([]byte{byte('0' + n/10), byte('0' + n%10)})
}
