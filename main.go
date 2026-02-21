package main

import (
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strings"
)

type BagOfWords map[string]int

var (
	hamBagOfWords  BagOfWords = make(BagOfWords)
	hamTotalCount  int
	spamBagOfWords BagOfWords = make(BagOfWords)
	spamTotalCount int
)

func tokenize(message string) []string {
	fields := strings.Fields(message)
	tokens := make([]string, len(fields))
	for i, field := range fields {
		tokens[i] = strings.ToUpper(field)
	}
	return tokens
}

func addFileToBag(path string, bagOfWords BagOfWords) error {
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

func calculateTotalCount(bagOfWords BagOfWords) int {
	total := 0
	for _, count := range bagOfWords {
		total += count
	}
	return total
}

func classifyFile(filePath string, hamBow, spamBow BagOfWords, hamTc, spamTc int) (hamProb, spamProb float64, err error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, 0, fmt.Errorf("could not read file %s: %w", filePath, err)
	}

	email := string(content)
	tokens := tokenize(email)

	logProbHam := 0.0
	logProbSpam := 0.0

	alpha := 1.0

	combinedVocab := make(map[string]struct{})
	for word := range hamBow {
		combinedVocab[word] = struct{}{}
	}
	for word := range spamBow {
		combinedVocab[word] = struct{}{}
	}
	vocabSize := float64(len(combinedVocab))
	if vocabSize == 0 {
		vocabSize = 1
	}

	for _, token := range tokens {
		countInHam := float64(hamBow[token])
		countInSpam := float64(spamBow[token])

		probTokenGivenHam := (countInHam + alpha) / (float64(hamTc) + alpha*vocabSize)
		probTokenGivenSpam := (countInSpam + alpha) / (float64(spamTc) + alpha*vocabSize)

		logProbHam += math.Log(probTokenGivenHam)
		logProbSpam += math.Log(probTokenGivenSpam)
	}

	return logProbHam, logProbSpam, nil
}

func classifyDirectory(dirPath string, hamBow, spamBow BagOfWords, hamTc, spamTc int) (spamCount, hamCount int, err error) {
	spamOutcomeCount := 0
	hamOutcomeCount := 0

	err = filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			logProbHam, logProbSpam, classifyErr := classifyFile(path, hamBow, spamBow, hamTc, spamTc)
			if classifyErr != nil {
				fmt.Fprintf(os.Stderr, "Error classifying file %s: %v\n", path, classifyErr)
				return nil
			}

			if logProbSpam > logProbHam {
				spamOutcomeCount++
			} else {
				hamOutcomeCount++
			}
		}
		return nil
	})

	return spamOutcomeCount, hamOutcomeCount, err
}

func main() {
	hamRoot := "./enron1/ham"
	spamRoot := "./enron1/spam"

	fmt.Println("Training on ham directory...")
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
	hamTotalCount = calculateTotalCount(hamBagOfWords)
	fmt.Printf("Finished training ham. Unique words: %d, Total words: %d\n", len(hamBagOfWords), hamTotalCount)

	fmt.Println("Training on spam directory...")
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
	spamTotalCount = calculateTotalCount(spamBagOfWords)
	fmt.Printf("Finished training spam. Unique words: %d, Total words: %d\n", len(spamBagOfWords), spamTotalCount)

	fmt.Println("--- Training Complete ---")

	fmt.Printf("\nClassifying HAM folder: %s\n", hamRoot)
	hamPredictedSpam, hamPredictedHam, err := classifyDirectory(hamRoot, hamBagOfWords, spamBagOfWords, hamTotalCount, spamTotalCount)
	if err != nil {
		fmt.Printf("Error classifying ham directory: %v\n", err)
		return
	}
	fmt.Printf("  Actual HAM files: Predicted SPAM: %d, Predicted HAM: %d\n", hamPredictedSpam, hamPredictedHam)

	fmt.Printf("\nClassifying SPAM folder: %s\n", spamRoot)
	spamPredictedSpam, spamPredictedHam, err := classifyDirectory(spamRoot, hamBagOfWords, spamBagOfWords, hamTotalCount, spamTotalCount)
	if err != nil {
		fmt.Printf("Error classifying spam directory: %v\n", err)
		return
	}
	fmt.Printf("  Actual SPAM files: Predicted SPAM: %d, Predicted HAM: %d\n", spamPredictedSpam, spamPredictedHam)
}
