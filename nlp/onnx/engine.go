//go:build onnx

// Package onnx implémente nlp.Engine avec un modèle BERT-NER au format ONNX
// via onnxruntime (cgo). Compilé uniquement avec `go build -tags onnx`.
//
// Fichiers attendus dans le dossier modèle : model_quantized.onnx (ou
// model.onnx), vocab.txt, config.json (id2label). Voir scripts/download-model.*
package onnx

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"

	ort "github.com/yalue/onnxruntime_go"

	"github.com/YoLaub/presidigo-go/nlp"
	"github.com/YoLaub/presidigo-go/nlp/bertner"
)

// maxSeqLen est la longueur maximale de séquence BERT ([CLS] + tokens + [SEP]).
const maxSeqLen = 512

// Engine exécute un modèle BERT-NER ONNX. Sûr pour un usage concurrent
// (la session onnxruntime est verrouillée par mutex).
type Engine struct {
	modelDir string
	libPath  string

	mu      sync.Mutex
	loaded  bool
	tok     *bertner.Tokenizer
	session *ort.DynamicAdvancedSession
	labels  []string
}

// Option configure le moteur.
type Option func(*Engine)

// WithLibrary fixe le chemin de la bibliothèque onnxruntime
// (onnxruntime.dll / libonnxruntime.so). Défaut : variable d'environnement
// ONNXRUNTIME_LIB, sinon le nom nu résolu par l'OS.
func WithLibrary(path string) Option {
	return func(e *Engine) { e.libPath = path }
}

// New crée le moteur pour un dossier modèle. Load() charge réellement.
func New(modelDir string, opts ...Option) *Engine {
	e := &Engine{modelDir: modelDir, libPath: os.Getenv("ONNXRUNTIME_LIB")}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Load initialise onnxruntime, le vocab et la session.
func (e *Engine) Load() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.loaded {
		return nil
	}

	if e.libPath != "" {
		ort.SetSharedLibraryPath(e.libPath)
	}
	if !ort.IsInitialized() {
		if err := ort.InitializeEnvironment(); err != nil {
			return fmt.Errorf("onnx: initialisation onnxruntime : %w", err)
		}
	}

	vocab, err := os.Open(filepath.Join(e.modelDir, "vocab.txt"))
	if err != nil {
		return fmt.Errorf("onnx: vocab.txt : %w", err)
	}
	defer vocab.Close()
	// bert-base-NER est un modèle CASED : la casse porte le signal des noms
	// propres. Minuscules uniquement si tokenizer_config.json le demande.
	if e.tok, err = bertner.NewTokenizer(vocab, doLowerCase(e.modelDir)); err != nil {
		return err
	}

	if e.labels, err = loadLabels(filepath.Join(e.modelDir, "config.json")); err != nil {
		return err
	}

	modelPath := filepath.Join(e.modelDir, "model_quantized.onnx")
	if _, err := os.Stat(modelPath); err != nil {
		modelPath = filepath.Join(e.modelDir, "model.onnx")
	}
	e.session, err = ort.NewDynamicAdvancedSession(modelPath,
		[]string{"input_ids", "attention_mask", "token_type_ids"},
		[]string{"logits"}, nil)
	if err != nil {
		return fmt.Errorf("onnx: session : %w", err)
	}
	e.loaded = true
	return nil
}

// Process tokenise le texte, exécute le modèle et agrège les labels BIO
// en entités NER (offsets en runes).
func (e *Engine) Process(_ context.Context, text, _ string) (*nlp.Artifacts, error) {
	if err := e.Load(); err != nil {
		return nil, err
	}
	tokens := e.tok.Tokenize(text)
	if len(tokens) == 0 {
		return &nlp.Artifacts{}, nil
	}
	if len(tokens) > maxSeqLen-2 {
		tokens = tokens[:maxSeqLen-2]
	}

	seq := len(tokens) + 2
	ids := make([]int64, seq)
	mask := make([]int64, seq)
	types := make([]int64, seq)
	ids[0] = int64(e.tok.ID("[CLS]"))
	for i, t := range tokens {
		ids[i+1] = int64(t.ID)
	}
	ids[seq-1] = int64(e.tok.ID("[SEP]"))
	for i := range mask {
		mask[i] = 1
	}

	logits, shape, err := e.run(ids, mask, types)
	if err != nil {
		return nil, err
	}
	numLabels := int(shape[2])

	labels := make([]bertner.TokenLabel, len(tokens))
	for i := range tokens {
		// +1 : décalage du [CLS] en tête de séquence.
		row := logits[(i+1)*numLabels : (i+2)*numLabels]
		idx, score := softmaxArgmax(row)
		if idx < len(e.labels) {
			labels[i] = bertner.TokenLabel{Label: e.labels[idx], Score: score}
		}
	}

	entities := bertner.Aggregate(tokens, labels)
	artifacts := &nlp.Artifacts{}
	for _, t := range tokens {
		artifacts.Tokens = append(artifacts.Tokens, t.Text)
	}
	for _, ent := range entities {
		artifacts.NerEntities = append(artifacts.NerEntities, nlp.NerEntity{
			Label: ent.Label, Start: ent.Start, End: ent.End, Score: ent.Score,
		})
	}
	return artifacts, nil
}

func (e *Engine) run(ids, mask, types []int64) ([]float32, []int64, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	seq := int64(len(ids))
	shape := ort.NewShape(1, seq)
	idsT, err := ort.NewTensor(shape, ids)
	if err != nil {
		return nil, nil, err
	}
	defer idsT.Destroy()
	maskT, err := ort.NewTensor(shape, mask)
	if err != nil {
		return nil, nil, err
	}
	defer maskT.Destroy()
	typesT, err := ort.NewTensor(shape, types)
	if err != nil {
		return nil, nil, err
	}
	defer typesT.Destroy()

	outputs := []ort.Value{nil}
	if err := e.session.Run([]ort.Value{idsT, maskT, typesT}, outputs); err != nil {
		return nil, nil, fmt.Errorf("onnx: inférence : %w", err)
	}
	out := outputs[0].(*ort.Tensor[float32])
	defer out.Destroy()

	data := make([]float32, len(out.GetData()))
	copy(data, out.GetData())
	return data, out.GetShape(), nil
}

// Destroy libère la session.
func (e *Engine) Destroy() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.session != nil {
		e.session.Destroy()
		e.session = nil
		e.loaded = false
	}
}

func softmaxArgmax(logits []float32) (int, float64) {
	maxIdx, maxVal := 0, float32(math.Inf(-1))
	for i, v := range logits {
		if v > maxVal {
			maxIdx, maxVal = i, v
		}
	}
	var sum float64
	for _, v := range logits {
		sum += math.Exp(float64(v - maxVal))
	}
	return maxIdx, 1.0 / sum
}

// doLowerCase lit do_lower_case du tokenizer_config.json (défaut : false,
// les modèles NER usuels étant cased).
func doLowerCase(modelDir string) bool {
	raw, err := os.ReadFile(filepath.Join(modelDir, "tokenizer_config.json"))
	if err != nil {
		return false
	}
	var cfg struct {
		DoLowerCase bool `json:"do_lower_case"`
	}
	if json.Unmarshal(raw, &cfg) != nil {
		return false
	}
	return cfg.DoLowerCase
}

// loadLabels lit id2label du config.json HuggingFace.
func loadLabels(path string) ([]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("onnx: config.json : %w", err)
	}
	var cfg struct {
		Id2Label map[string]string `json:"id2label"`
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("onnx: config.json : %w", err)
	}
	type kv struct {
		id    int
		label string
	}
	pairs := make([]kv, 0, len(cfg.Id2Label))
	for k, v := range cfg.Id2Label {
		id, err := strconv.Atoi(k)
		if err != nil {
			return nil, fmt.Errorf("onnx: id2label : %w", err)
		}
		pairs = append(pairs, kv{id, v})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].id < pairs[j].id })
	labels := make([]string, len(pairs))
	for i, p := range pairs {
		labels[i] = p.label
	}
	return labels, nil
}
