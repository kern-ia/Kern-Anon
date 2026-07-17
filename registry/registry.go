// Package registry gère l'ensemble des recognizers actifs et leur filtrage
// par langue et par entité.
package registry

import (
	"slices"
	"sync"

	"github.com/YoLaub/PresidioGo/recognizer"
)

// Registry contient les recognizers enregistrés. Sûr pour un usage concurrent.
type Registry struct {
	mu          sync.RWMutex
	recognizers []recognizer.Recognizer
}

// New crée un registry vide.
func New() *Registry {
	return &Registry{}
}

// Add enregistre un recognizer.
func (r *Registry) Add(rec recognizer.Recognizer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recognizers = append(r.recognizers, rec)
}

// Remove retire le recognizer portant ce nom.
func (r *Registry) Remove(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recognizers = slices.DeleteFunc(r.recognizers, func(rec recognizer.Recognizer) bool {
		return rec.Name() == name
	})
}

// Get retourne les recognizers de la langue donnée. Si des entités sont
// précisées, seuls les recognizers en supportant au moins une sont retournés.
func (r *Registry) Get(language string, entities ...string) []recognizer.Recognizer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []recognizer.Recognizer
	for _, rec := range r.recognizers {
		if rec.Language() != language {
			continue
		}
		if len(entities) > 0 && !supportsAny(rec, entities) {
			continue
		}
		out = append(out, rec)
	}
	return out
}

// SupportedEntities retourne les entités distinctes supportées pour une langue.
func (r *Registry) SupportedEntities(language string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []string
	for _, rec := range r.recognizers {
		if rec.Language() != language {
			continue
		}
		for _, e := range rec.SupportedEntities() {
			if !slices.Contains(out, e) {
				out = append(out, e)
			}
		}
	}
	return out
}

func supportsAny(rec recognizer.Recognizer, entities []string) bool {
	for _, e := range rec.SupportedEntities() {
		if slices.Contains(entities, e) {
			return true
		}
	}
	return false
}
