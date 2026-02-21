package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var hamBagOfWords = make(map[string]int)
var hamTotalCount int

var spamBagOfWords = make(map[string]int)
var spamTotalCount int

func tokenize(message string) []string {
	fields := strings.Fields(message)
	tokens := make([]string, len(fields))
	for i, field := range fields {
		tokens[i] = strings.ToUpper(field)
	}
	return tokens
}

func addFileToBag(path string, bagOfWords map[string]int) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read file %s: %w", path, err)
	}

	emailContent := string(content)
	tokens := tokenize(emailContent)

	for _, token := range tokens {
		bagOfWords[token]++
	}
	return nil
}

func main() {
	hamRoot := "./enron1/ham"
	spamRoot := "./enron1/spam"

	fmt.Println("Processing ham directory...")
	err := filepath.WalkDir(hamRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			fileErr := addFileToBag(path, hamBagOfWords)
			if fileErr != nil {
				fmt.Fprintf(os.Stderr, "Error processing ham file %s: %v\n", path, fileErr)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking ham directory: %v\n", err)
		return
	}

	for _, count := range hamBagOfWords {
		hamTotalCount += count
	}
	fmt.Printf("Finished processing ham directory. Total unique ham words: %d, Total ham word count: %d\n",
		len(hamBagOfWords), hamTotalCount)

	fmt.Println("Processing spam directory...")
	err = filepath.WalkDir(spamRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			fileErr := addFileToBag(path, spamBagOfWords)
			if fileErr != nil {
				fmt.Fprintf(os.Stderr, "Error processing spam file %s: %v\n", path, fileErr)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking spam directory: %v\n", err)
		return
	}

	for _, count := range spamBagOfWords {
		spamTotalCount += count
	}
	fmt.Printf("Finished processing spam directory. Total unique spam words: %d, Total spam word count: %d\n",
		len(spamBagOfWords), spamTotalCount)
}
