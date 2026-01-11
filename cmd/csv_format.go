package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// escapeExecutionPlan converts actual newlines to literal \n for CSV storage
func escapeExecutionPlan(plan string) string {
	return strings.ReplaceAll(plan, "\n", "\\n")
}

// createCSVWriter creates a standardized CSV writer with proper settings
func createCSVWriter(file *os.File) *csv.Writer {
	writer := csv.NewWriter(file)
	writer.Comma = ','
	writer.UseCRLF = false // Unix-style line endings
	return writer
}

// writeCSVPlan generates a CSV file for analyze command
// Returns absolute path of generated file
func writeCSVPlan(plan, query, title string, costInfo *CostInfo) string {
	fileName := title + ".csv"

	file, err := os.Create(fileName)
	if err != nil {
		logErrorAndExit("unable to create CSV file: ", err)
	}
	defer file.Close()

	writer := createCSVWriter(file)
	defer writer.Flush()

	// Write header row
	header := []string{
		"title",
		"query",
		"execution_plan",
		"total_cost",
		"exceeds_threshold",
		"threshold_value",
		"expensive_ops_count",
		"generated_at",
	}
	if err := writer.Write(header); err != nil {
		logErrorAndExit("unable to write CSV header: ", err)
	}

	// Prepare data row
	totalCost := "N/A"
	exceedsThreshold := "false"
	thresholdValue := "0"
	expensiveOpsCount := "0"

	if costInfo != nil {
		totalCost = fmt.Sprintf("%.2f", costInfo.TotalCost)
		exceedsThreshold = strconv.FormatBool(costInfo.ExceedsLimit)
		thresholdValue = fmt.Sprintf("%.2f", costInfo.ThresholdValue)
		expensiveOpsCount = strconv.Itoa(len(costInfo.ExpensiveOps))
	}

	row := []string{
		title,
		query,
		escapeExecutionPlan(plan),
		totalCost,
		exceedsThreshold,
		thresholdValue,
		expensiveOpsCount,
		time.Now().Format(time.RFC3339),
	}

	if err := writer.Write(row); err != nil {
		logErrorAndExit("unable to write CSV data: ", err)
	}

	abs, err := filepath.Abs(file.Name())
	if err != nil {
		logErrorAndExit("unable to get file absolute path: ", err)
	}

	return abs
}

// writeComparisonCSV generates a CSV file for compare command
// Returns absolute path of generated file
func writeComparisonCSV(result *ComparisonResult) {
	title := generateTitle()
	fileName := fmt.Sprintf("Comparison_%s.csv", title)

	file, err := os.Create(fileName)
	if err != nil {
		logErrorAndExit("unable to create CSV file: ", err)
	}
	defer file.Close()

	writer := createCSVWriter(file)
	defer writer.Flush()

	// Write header row
	header := []string{
		"query1",
		"query2",
		"plan1",
		"plan2",
		"cost1",
		"cost2",
		"winner",
		"cost_diff",
		"cost_diff_pct",
		"recommendation",
	}
	if err := writer.Write(header); err != nil {
		logErrorAndExit("unable to write CSV header: ", err)
	}

	// Prepare data row
	row := []string{
		result.Query1,
		result.Query2,
		escapeExecutionPlan(result.Plan1),
		escapeExecutionPlan(result.Plan2),
		fmt.Sprintf("%.2f", result.Cost1.TotalCost),
		fmt.Sprintf("%.2f", result.Cost2.TotalCost),
		result.Winner,
		fmt.Sprintf("%.2f", result.CostDiff),
		fmt.Sprintf("%.2f", result.CostDiffPct),
		result.Recommendation,
	}

	if err := writer.Write(row); err != nil {
		logErrorAndExit("unable to write CSV data: ", err)
	}

	abs, err := filepath.Abs(file.Name())
	if err != nil {
		logErrorAndExit("unable to get file absolute path: ", err)
	}

	// Display success message
	fmt.Println("\n" + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📁 Comparison saved successfully!")
	fmt.Printf("   %s\n", abs)
	fmt.Printf("\n🏆 Winner: %s\n", result.Winner)
	if result.CostDiff != 0 {
		fmt.Printf("Cost Difference: %.2f (%.2f%%)\n", result.CostDiff, result.CostDiffPct)
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
}

// writeCSVBatchReport generates a CSV file for batch command (combined mode)
// Returns absolute path of generated file
func writeCSVBatchReport(report BatchReport, fileName string) string {
	csvFileName := fileName + ".csv"

	file, err := os.Create(csvFileName)
	if err != nil {
		logErrorAndExit("unable to create CSV file: ", err)
	}
	defer file.Close()

	writer := createCSVWriter(file)
	defer writer.Flush()

	// Write header row
	header := []string{
		"query_number",
		"query",
		"execution_plan",
		"total_cost",
		"exceeds_threshold",
		"error",
		"status",
		"generated_at",
	}
	if err := writer.Write(header); err != nil {
		logErrorAndExit("unable to write CSV header: ", err)
	}

	// Write data rows (one per query)
	for _, result := range report.Results {
		status := "success"
		if result.Error != "" {
			status = "failed"
		}

		totalCost := ""
		exceedsThreshold := "false"
		executionPlan := ""

		if result.Error == "" {
			if result.CostAnalysis != nil {
				totalCost = fmt.Sprintf("%.2f", result.CostAnalysis.TotalCost)
				exceedsThreshold = strconv.FormatBool(result.CostAnalysis.ExceedsLimit)
			}
			executionPlan = escapeExecutionPlan(result.ExecutionPlan)
		}

		row := []string{
			strconv.Itoa(result.QueryNumber),
			result.Query,
			executionPlan,
			totalCost,
			exceedsThreshold,
			result.Error,
			status,
			result.GeneratedAt.Format(time.RFC3339),
		}

		if err := writer.Write(row); err != nil {
			logErrorAndExit("unable to write CSV data: ", err)
		}
	}

	abs, err := filepath.Abs(file.Name())
	if err != nil {
		logErrorAndExit("unable to get file absolute path: ", err)
	}

	return abs
}
