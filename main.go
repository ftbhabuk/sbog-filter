package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func frequency(tokens []string) map[string]int {
	freqs := make(map[string]int)
	for _, token := range tokens {
		freqs[token]++
	}
	return freqs
}

func probability(freqs map[string]int, total int) map[string]float64 {
	probs := make(map[string]float64)
	for word, count := range freqs {
		probs[word] = float64(count) / float64(total)
	}
	return probs
}

func AddFileToBag(path string, bowl map[string]int) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read file %s: %w", path, err)
	}

	email := string(content)
	tokens := strings.Fields(email)

	for _, token := range tokens {
		bowl[strings.ToUpper(token)]++
	}
	return nil
}

func main() {
	bowl := make(map[string]int)

	err := filepath.WalkDir("./enron1", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		return AddFileToBag(path, bowl)
	})
	if err != nil {
		panic(err)
	}

	freqs := bowl
	total := 0
	for _, count := range freqs {
		total += count
	}
	probs := probability(freqs, total)

	type pair struct {
		word  string
		count int
		prob  float64
	}
	var sorted []pair
	for w, c := range freqs {
		sorted = append(sorted, pair{w, c, probs[w]})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})

	fmt.Printf("Total words: %d\n\n", total)
	fmt.Printf("%-15s %6s %10s\n", "Word", "Count", "Probability")
	fmt.Println("--------------------------------------")
	for _, p := range sorted[:20] {
		fmt.Printf("%-15s %6d %10.6f\n", p.word, p.count, p.prob)
	}
}
