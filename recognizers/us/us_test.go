package us_test

import (
	"testing"

	"github.com/YoLaub/PresidioGo/internal/oracletest"
	"github.com/YoLaub/PresidioGo/recognizers/us"
)

func TestOracle(t *testing.T) {
	recs := us.All("en")
	if len(recs) != 9 {
		t.Fatalf("us.All doit fournir 9 recognizers, obtenu %d", len(recs))
	}
	oracletest.Run(t, recs)
}

func TestAbaChecksum(t *testing.T) {
	if !us.AbaChecksum("122105155") {
		t.Error("122105155 devrait être un routing number valide")
	}
	if us.AbaChecksum("122105156") {
		t.Error("122105156 devrait être invalide")
	}
}

func TestNpiChecksum(t *testing.T) {
	if !us.NpiChecksum("1234567893") {
		t.Error("1234567893 devrait être un NPI valide (Luhn préfixé 80840)")
	}
	if us.NpiChecksum("1234567890") {
		t.Error("1234567890 devrait être invalide")
	}
	// Dégénéré : corps de chiffres identiques rejeté (invalidate du fork).
	if us.NpiChecksum("1111111112") {
		t.Error("corps dégénéré 111111111x : rejet attendu")
	}
}

func TestDeaChecksum(t *testing.T) {
	if !us.DeaChecksum("AB1234563") {
		t.Error("AB1234563 devrait être valide (somme DEA = 33 → 3)")
	}
	if us.DeaChecksum("AB1234564") {
		t.Error("AB1234564 devrait être invalide")
	}
}
