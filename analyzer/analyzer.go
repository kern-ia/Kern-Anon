// Package analyzer orchestre le pipeline de détection : moteur NLP →
// recognizers du registry → boost contextuel → dédoublonnage → seuil de score.
// Équivalent Go de l'AnalyzerEngine du fork.
package analyzer

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"sync"

	"github.com/YoLaub/presidigo-go/contextaware"
	"github.com/YoLaub/presidigo-go/nlp"
	"github.com/YoLaub/presidigo-go/pii"
	"github.com/YoLaub/presidigo-go/recognizer"
	"github.com/YoLaub/presidigo-go/registry"
)

// Engine est le moteur d'analyse.
type Engine struct {
	registry        *registry.Registry
	nlpEngine       nlp.Engine
	enhancer        *contextaware.Enhancer
	defaultLanguage string
}

// Option configure le moteur à la construction.
type Option func(*Engine)

// WithRegistry fournit le registry de recognizers (obligatoire).
func WithRegistry(r *registry.Registry) Option {
	return func(e *Engine) { e.registry = r }
}

// WithNlpEngine branche un moteur NLP (nlp.NoOp{} ou l'ONNX derrière le tag).
func WithNlpEngine(n nlp.Engine) Option {
	return func(e *Engine) { e.nlpEngine = n }
}

// WithEnhancer remplace l'enhancer contextuel (nil pour le désactiver).
func WithEnhancer(en *contextaware.Enhancer) Option {
	return func(e *Engine) { e.enhancer = en }
}

// WithDefaultLanguage change la langue par défaut ("en").
func WithDefaultLanguage(lang string) Option {
	return func(e *Engine) { e.defaultLanguage = lang }
}

// New construit le moteur. Un registry est requis ; l'enhancer contextuel
// est activé par défaut avec les paramètres du fork.
func New(opts ...Option) (*Engine, error) {
	e := &Engine{
		enhancer:        contextaware.New(),
		defaultLanguage: "en",
	}
	for _, opt := range opts {
		opt(e)
	}
	if e.registry == nil {
		return nil, errors.New("analyzer: un registry est requis (WithRegistry)")
	}
	if e.nlpEngine != nil {
		if err := e.nlpEngine.Load(); err != nil {
			return nil, fmt.Errorf("analyzer: chargement du moteur NLP : %w", err)
		}
	}
	return e, nil
}

type callConfig struct {
	language string
	entities []string
	minScore float64
}

// CallOption configure un appel à Analyze.
type CallOption func(*callConfig)

// Language fixe la langue de l'analyse (défaut : langue par défaut du moteur).
func Language(lang string) CallOption {
	return func(c *callConfig) { c.language = lang }
}

// Entities restreint l'analyse à ces types d'entités.
func Entities(entities ...string) CallOption {
	return func(c *callConfig) { c.entities = entities }
}

// MinScore écarte les résultats sous ce seuil.
func MinScore(score float64) CallOption {
	return func(c *callConfig) { c.minScore = score }
}

// Analyze détecte les entités PII du texte et retourne les résultats triés
// par position (offsets en runes).
func (e *Engine) Analyze(ctx context.Context, text string, opts ...CallOption) ([]pii.Result, error) {
	cfg := callConfig{language: e.defaultLanguage}
	for _, opt := range opts {
		opt(&cfg)
	}

	var artifacts *nlp.Artifacts
	if e.nlpEngine != nil {
		var err error
		if artifacts, err = e.nlpEngine.Process(ctx, text, cfg.language); err != nil {
			return nil, fmt.Errorf("analyzer: moteur NLP : %w", err)
		}
	}

	// Les recognizers sont indépendants et thread-safe : fan-out en
	// goroutines, résultats réassemblés dans l'ordre du registry
	// (déterminisme conservé).
	recognizers := e.registry.Get(cfg.language, cfg.entities...)
	perRec := make([][]pii.Result, len(recognizers))
	errs := make([]error, len(recognizers))
	var wg sync.WaitGroup
	for i, rec := range recognizers {
		wg.Add(1)
		go func(i int, rec recognizer.Recognizer) {
			defer wg.Done()
			perRec[i], errs[i] = rec.Analyze(ctx, text, artifacts)
		}(i, rec)
	}
	wg.Wait()
	var results []pii.Result
	for i, rec := range recognizers {
		if errs[i] != nil {
			return nil, fmt.Errorf("analyzer: %s : %w", rec.Name(), errs[i])
		}
		results = append(results, perRec[i]...)
	}

	if len(cfg.entities) > 0 {
		results = slices.DeleteFunc(results, func(r pii.Result) bool {
			return !slices.Contains(cfg.entities, r.EntityType)
		})
	}

	if e.enhancer != nil {
		results = e.enhancer.Enhance(text, results, recognizers)
	}

	results = removeContainedDuplicates(results)

	if cfg.minScore > 0 {
		results = slices.DeleteFunc(results, func(r pii.Result) bool {
			return r.Score < cfg.minScore
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Start != results[j].Start {
			return results[i].Start < results[j].Start
		}
		return results[i].End > results[j].End
	})
	return results, nil
}

// removeContainedDuplicates écarte tout résultat contenu dans un résultat du
// MÊME type d'entité de score supérieur ou égal (sémantique du fork : les
// chevauchements entre entités différentes sont conservés — c'est
// l'anonymizer qui les tranche).
func removeContainedDuplicates(results []pii.Result) []pii.Result {
	byPriority := make([]pii.Result, len(results))
	copy(byPriority, results)
	sort.SliceStable(byPriority, func(i, j int) bool {
		a, b := byPriority[i], byPriority[j]
		if a.Score != b.Score {
			return a.Score > b.Score
		}
		return a.End-a.Start > b.End-b.Start
	})

	var kept []pii.Result
	for _, cand := range byPriority {
		contained := false
		for _, k := range kept {
			if k.EntityType == cand.EntityType &&
				k.Start <= cand.Start && cand.End <= k.End {
				contained = true
				break
			}
		}
		if !contained {
			kept = append(kept, cand)
		}
	}
	return kept
}
