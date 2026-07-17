package registry

import (
	"github.com/YoLaub/PresidioGo/recognizers/fr"
	"github.com/YoLaub/PresidioGo/recognizers/generic"
	"github.com/YoLaub/PresidioGo/recognizers/us"
)

// Default retourne un registry pré-rempli pour les langues demandées :
// les recognizers génériques pour chaque langue, plus les locales
// disponibles ("en" → US, "fr" → France).
func Default(languages ...string) *Registry {
	reg := New()
	for _, lang := range languages {
		for _, rec := range generic.All(lang) {
			reg.Add(rec)
		}
		switch lang {
		case "en":
			for _, rec := range us.All(lang) {
				reg.Add(rec)
			}
		case "fr":
			for _, rec := range fr.All() {
				reg.Add(rec)
			}
		}
	}
	return reg
}
