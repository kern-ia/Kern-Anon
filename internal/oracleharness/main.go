// Le harness oracle compare les sorties du moteur Go à celles du
// presidio-analyzer Python du fork sur le corpus oracle.
//
// Usage :
//
//	docker run -d -p 5002:5001 -e PORT=5001 <image presidio-analyzer>
//	go run ./internal/oracleharness            # PRESIDIO_URL défaut http://localhost:5002
//
// Critère v0.1.0 (docs/PLAN.md §7) : ≥ 95 % d'accord sur les entités pattern.
// Le NER n'est pas comparé (moteurs différents : spaCy vs BERT ONNX).
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/YoLaub/presidigo-go/analyzer"
	"github.com/YoLaub/presidigo-go/registry"
)

// excluded : entités hors comparaison — NER (moteurs différents : spaCy vs
// BERT), DATE_TIME/PHONE_NUMBER (recognizers non portés), et ABA/NPI/MBI que
// le registre PAR DÉFAUT du service Python ne charge pas (voir /recognizers).
var excluded = map[string]bool{
	"PERSON": true, "LOCATION": true, "ORGANIZATION": true, "NRP": true,
	"DATE_TIME": true, "PHONE_NUMBER": true,
	"ABA_ROUTING_NUMBER": true, "US_NPI": true, "US_MBI": true,
}

type oracleCase struct {
	ID       string `json:"id"`
	Language string `json:"language"`
	Text     string `json:"text"`
}

type pyResult struct {
	EntityType string  `json:"entity_type"`
	Start      int     `json:"start"`
	End        int     `json:"end"`
	Score      float64 `json:"score"`
}

func main() {
	url := os.Getenv("PRESIDIO_URL")
	if url == "" {
		url = "http://localhost:5002"
	}
	if err := waitHealthy(url); err != nil {
		fmt.Fprintf(os.Stderr, "presidio-analyzer injoignable sur %s : %v\n", url, err)
		os.Exit(1)
	}

	cases := loadCases()
	eng, err := analyzer.New(analyzer.WithRegistry(registry.Default("en", "fr")))
	if err != nil {
		panic(err)
	}

	agree, total := 0, 0
	for _, c := range cases {
		if c.Text == "" {
			continue
		}
		lang := c.Language
		if lang != "en" { // le service Python du fork ne charge que l'anglais
			continue
		}
		goResults, err := eng.Analyze(context.Background(), c.Text, analyzer.Language(lang))
		if err != nil {
			panic(err)
		}
		pyResults := analyzePython(url, c.Text, lang)

		// Comparaison sur l'union des spans pattern détectés de part et d'autre.
		type span struct {
			entity     string
			start, end int
		}
		goSet := map[span]bool{}
		for _, r := range goResults {
			if !excluded[r.EntityType] && r.Score >= 0.4 {
				goSet[span{r.EntityType, r.Start, r.End}] = true
			}
		}
		pySet := map[span]bool{}
		for _, r := range pyResults {
			if !excluded[r.EntityType] && r.Score >= 0.4 {
				pySet[span{r.EntityType, r.Start, r.End}] = true
			}
		}
		for s := range goSet {
			total++
			if pySet[s] {
				agree++
			} else {
				fmt.Printf("  [%s] Go seul   : %s [%d:%d]\n", c.ID, s.entity, s.start, s.end)
			}
		}
		for s := range pySet {
			if !goSet[s] {
				total++
				fmt.Printf("  [%s] Python seul : %s [%d:%d]\n", c.ID, s.entity, s.start, s.end)
			}
		}
	}

	pct := 100.0
	if total > 0 {
		pct = float64(agree) / float64(total) * 100
	}
	fmt.Printf("\nAccord Go/Python sur les entités pattern : %d/%d (%.1f %%)\n", agree, total, pct)
	if pct < 95 {
		fmt.Println("SOUS le critère v0.1.0 (95 %)")
		os.Exit(1)
	}
	fmt.Println("Critère v0.1.0 atteint (>= 95 %)")
}

func loadCases() []oracleCase {
	f, err := os.Open("internal/testdata/oracle.jsonl")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var cases []oracleCase
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if len(sc.Bytes()) == 0 {
			continue
		}
		var c oracleCase
		if err := json.Unmarshal(sc.Bytes(), &c); err != nil {
			panic(err)
		}
		cases = append(cases, c)
	}
	return cases
}

func waitHealthy(url string) error {
	var lastErr error
	for i := 0; i < 30; i++ {
		resp, err := http.Get(url + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		lastErr = err
		time.Sleep(2 * time.Second)
	}
	return lastErr
}

func analyzePython(url, text, lang string) []pyResult {
	body, _ := json.Marshal(map[string]any{"text": text, "language": lang})
	resp, err := http.Post(url+"/analyze", "application/json", bytes.NewReader(body))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var results []pyResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		panic(err)
	}
	return results
}
