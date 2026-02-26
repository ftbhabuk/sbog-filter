package main

import (
	"encoding/csv"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strconv"
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

func saveModelToCSV(hamPath, spamPath string, hamBow, spamBow BagOfWords, hamTc, spamTc int) error {
	hamFile, err := os.Create(hamPath)
	if err != nil {
		return err
	}
	defer hamFile.Close()

	spamFile, err := os.Create(spamPath)
	if err != nil {
		return err
	}
	defer spamFile.Close()

	hamWriter := csv.NewWriter(hamFile)
	spamWriter := csv.NewWriter(spamFile)

	hamWriter.Write([]string{"word", "count"})
	for word, count := range hamBow {
		hamWriter.Write([]string{word, strconv.Itoa(count)})
	}
	hamWriter.Flush()

	spamWriter.Write([]string{"word", "count"})
	for word, count := range spamBow {
		spamWriter.Write([]string{word, strconv.Itoa(count)})
	}
	spamWriter.Flush()

	hamMeta, _ := os.Create("ham_meta.txt")
	hamMeta.WriteString(strconv.Itoa(hamTc))
	hamMeta.Close()

	spamMeta, _ := os.Create("spam_meta.txt")
	spamMeta.WriteString(strconv.Itoa(spamTc))
	spamMeta.Close()

	return nil
}

func loadModelFromCSV(hamPath, spamPath string) (BagOfWords, BagOfWords, int, int, error) {
	hamBow := make(BagOfWords)
	spamBow := make(BagOfWords)

	hamFile, err := os.Open(hamPath)
	if err != nil {
		return hamBow, spamBow, 0, 0, err
	}
	defer hamFile.Close()

	hamReader := csv.NewReader(hamFile)
	records, _ := hamReader.ReadAll()
	for i, row := range records {
		if i == 0 {
			continue
		}
		if len(row) >= 2 {
			count, _ := strconv.Atoi(row[1])
			hamBow[row[0]] = count
		}
	}

	spamFile, err := os.Open(spamPath)
	if err != nil {
		return hamBow, spamBow, 0, 0, err
	}
	defer spamFile.Close()

	spamReader := csv.NewReader(spamFile)
	records, _ = spamReader.ReadAll()
	for i, row := range records {
		if i == 0 {
			continue
		}
		if len(row) >= 2 {
			count, _ := strconv.Atoi(row[1])
			spamBow[row[0]] = count
		}
	}

	hamTc := 0
	spamTc := 0

	if data, err := os.ReadFile("ham_meta.txt"); err == nil {
		hamTc, _ = strconv.Atoi(strings.TrimSpace(string(data)))
	}
	if data, err := os.ReadFile("spam_meta.txt"); err == nil {
		spamTc, _ = strconv.Atoi(strings.TrimSpace(string(data)))
	}

	return hamBow, spamBow, hamTc, spamTc, nil
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
	hamBagOfWords, spamBagOfWords, hamTotalCount, spamTotalCount, err := loadModelFromCSV("data/ham.csv", "data/spam.csv")
	if err == nil {
		fmt.Println("--- Loaded cached model from CSV ---")
		fmt.Printf("Loaded: %d unique ham words, %d unique spam words\n", len(hamBagOfWords), len(spamBagOfWords))
		fmt.Printf("Ham total: %d, Spam total: %d\n", hamTotalCount, spamTotalCount)
	} else {
		trainingFolders := []string{
			"./data/enron1",
			"./data/enron2",
			"./data/enron3",
			"./data/enron4",
			"./data/enron5",
		}

		fmt.Println("--- Starting Training ---")
		for _, folder := range trainingFolders {
			hamPath := filepath.Join(folder, "ham")
			spamPath := filepath.Join(folder, "spam")

			fmt.Printf("Processing ham in %s...\n", folder)
			err := filepath.WalkDir(hamPath, func(path string, d fs.DirEntry, err error) error {
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
				fmt.Printf("Error walking ham directory %s: %v\n", hamPath, err)
				return
			}

			fmt.Printf("Processing spam in %s...\n", folder)
			err = filepath.WalkDir(spamPath, func(path string, d fs.DirEntry, err error) error {
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
				fmt.Printf("Error walking spam directory %s: %v\n", spamPath, err)
				return
			}
		}

		hamTotalCount = calculateTotalCount(hamBagOfWords)
		spamTotalCount = calculateTotalCount(spamBagOfWords)
		fmt.Printf("Finished training. Total unique ham words: %d, Total ham words: %d\n", len(hamBagOfWords), hamTotalCount)
		fmt.Printf("Finished training. Total unique spam words: %d, Total spam words: %d\n", len(spamBagOfWords), spamTotalCount)
		fmt.Println("--- Training Complete ---")

		err = saveModelToCSV("data/ham.csv", "data/spam.csv", hamBagOfWords, spamBagOfWords, hamTotalCount, spamTotalCount)
		if err == nil {
			fmt.Println("Model saved to data/ham.csv and data/spam.csv")
		}
	}

	// Test on enron6
	validationHamDir := "./data/enron6/ham"
	validationSpamDir := "./data/enron6/spam"

	fmt.Printf("\nClassifying HAM folder (enron6): %s\n", validationHamDir)
	hamPredictedSpam, hamPredictedHam, err := classifyDirectory(validationHamDir, hamBagOfWords, spamBagOfWords, hamTotalCount, spamTotalCount)
	if err != nil {
		fmt.Printf("Error classifying ham directory: %v\n", err)
		return
	}
	fmt.Printf("  Actual HAM files: Predicted SPAM: %d, Predicted HAM: %d\n", hamPredictedSpam, hamPredictedHam)

	fmt.Printf("\nClassifying SPAM folder (enron6): %s\n", validationSpamDir)
	spamPredictedSpam, spamPredictedHam, err := classifyDirectory(validationSpamDir, hamBagOfWords, spamBagOfWords, hamTotalCount, spamTotalCount)
	if err != nil {
		fmt.Printf("Error classifying spam directory: %v\n", err)
		return
	}
	fmt.Printf("  Actual SPAM files: Predicted SPAM: %d, Predicted HAM: %d\n", spamPredictedSpam, spamPredictedHam)

	printSummaryReport(hamPredictedSpam, hamPredictedHam, spamPredictedSpam, spamPredictedHam)
}
