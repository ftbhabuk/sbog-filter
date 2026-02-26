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

func calculateMetrics(truePositives, falsePositives, trueNegatives, falseNegatives int) (accuracy, precision, recall, f1 float64) {
	accuracy = float64(truePositives+trueNegatives) / float64(truePositives+falsePositives+trueNegatives+falseNegatives)
	precision = float64(truePositives) / float64(truePositives+falsePositives)
	recall = float64(truePositives) / float64(truePositives+falseNegatives)
	f1 = 2 * (precision * recall) / (precision + recall)
	return
}

func printSummaryReport(hamPredictedSpam, hamPredictedHam, spamPredictedSpam, spamPredictedHam int) {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("           CLASSIFICATION SUMMARY REPORT")
	fmt.Println(strings.Repeat("=", 50))

	truePositives := spamPredictedSpam
	falsePositives := hamPredictedSpam
	trueNegatives := hamPredictedHam
	falseNegatives := spamPredictedHam

	totalFiles := truePositives + falsePositives + trueNegatives + falseNegatives

	fmt.Printf("\n  Confusion Matrix:\n")
	fmt.Printf("                    Predicted\n")
	fmt.Printf("                 Spam      Ham\n")
	fmt.Printf("  Actual Spam   %5d    %5d\n", truePositives, falseNegatives)
	fmt.Printf("  Actual Ham    %5d    %5d\n", falsePositives, trueNegatives)

	accuracy, precision, recall, f1 := calculateMetrics(truePositives, falsePositives, trueNegatives, falseNegatives)

	fmt.Printf("\n  Metrics:\n")
	fmt.Printf("    Total Files:     %d\n", totalFiles)
	fmt.Printf("    Accuracy:        %.2f%%\n", accuracy*100)
	fmt.Printf("    Precision:       %.2f%%\n", precision*100)
	fmt.Printf("    Recall:          %.2f%%\n", recall*100)
	fmt.Printf("    F1 Score:        %.2f%%\n", f1*100)

	fmt.Printf("\n  Breakdown:\n")
	fmt.Printf("    Ham correctly classified:    %d / %d (%.1f%%)\n", trueNegatives, trueNegatives+falsePositives, float64(trueNegatives)/float64(trueNegatives+falsePositives)*100)
	fmt.Printf("    Spam correctly classified:   %d / %d (%.1f%%)\n", truePositives, truePositives+falseNegatives, float64(truePositives)/float64(truePositives+falseNegatives)*100)

	printAsciiBarChart(hamPredictedHam, hamPredictedSpam, spamPredictedHam, spamPredictedSpam)

	fmt.Println(strings.Repeat("=", 50))
}

func printAsciiBarChart(hamCorrect, hamWrong, spamCorrect, spamWrong int) {
	const barWidth = 30

	hamTotal := hamCorrect + hamWrong
	spamTotal := spamCorrect + spamWrong

	hamPct := float64(hamCorrect) / float64(hamTotal) * 100
	spamPct := float64(spamCorrect) / float64(spamTotal) * 100

	hamBars := int(hamPct / 100 * float64(barWidth))
	spamBars := int(spamPct / 100 * float64(barWidth))

	fmt.Printf("\n  ASCII Bar Chart (accuracy per category):\n\n")
	fmt.Printf("    HAM (%d/%d):\n", hamCorrect, hamTotal)
	fmt.Printf("    [")
	for i := 0; i < barWidth; i++ {
		if i < hamBars {
			fmt.Print("█")
		} else {
			fmt.Print("░")
		}
	}
	fmt.Printf("] %.1f%%\n", hamPct)

	fmt.Printf("    SPAM (%d/%d):\n", spamCorrect, spamTotal)
	fmt.Printf("    [")
	for i := 0; i < barWidth; i++ {
		if i < spamBars {
			fmt.Print("█")
		} else {
			fmt.Print("░")
		}
	}
	fmt.Printf("] %.1f%%\n", spamPct)
}

func main() {
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
