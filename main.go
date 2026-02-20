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

func main() {
	content, err := ioutil.ReadFile("./enron1/ham/0002.1999-12-13.farmer.ham.txt")
	if err != nil {
		panic(err)
	}
	tokens := tokenize(string(content))
	freqs := frequency(tokens)

	type pair struct {
		word  string
		count int
	}
	var sorted []pair
	for w, c := range freqs {
		sorted = append(sorted, pair{w, c})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})

	for _, p := range sorted[:20] {
		fmt.Printf("%s: %d\n", p.word, p.count)
	}
}
