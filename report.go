package main

import (
	"fmt"
	"strings"
)

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
