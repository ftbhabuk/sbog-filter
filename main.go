package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"sort"
	"strings"
)

func tokenize(text string) []string {
	re := regexp.MustCompile(`[a-zA-Z]+`)
	matches := re.FindAllString(strings.ToLower(text), -1)
	return matches
}

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

func main() {
	content, err := ioutil.ReadFile("./enron1/ham/0002.1999-12-13.farmer.ham.txt")
	if err != nil {
		panic(err)
	}
	tokens := tokenize(string(content))
	freqs := frequency(tokens)
	probs := probability(freqs, len(tokens))

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

	fmt.Printf("Total words: %d\n\n", len(tokens))
	fmt.Printf("%-15s %6s %10s\n", "Word", "Count", "Probability")
	fmt.Println("--------------------------------------")
	for _, p := range sorted[:20] {
		fmt.Printf("%-15s %6d %10.6f\n", p.word, p.count, p.prob)
	}
}
