package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// escapeMarkdownSpecialChars escapes special markdown characters in text
func escapeMarkdownSpecialChars(text string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`*`, `\*`,
		`_`, `\_`,
		`[`, `\[`,
		`]`, `\]`,
	)
	return replacer.Replace(text)
}

// formatCostInfoMarkdown formats cost analysis information as markdown
func formatCostInfoMarkdown(costInfo *CostInfo) string {
	if costInfo == nil {
		return "_Cost analysis not available (threshold not set)_\n"
	}

	var sb strings.Builder

	// Cost analysis table
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Total Cost | %.2f |\n", costInfo.TotalCost))
	sb.WriteString(fmt.Sprintf("| Exceeds Threshold | %t |\n", costInfo.ExceedsLimit))
	sb.WriteString(fmt.Sprintf("| Threshold Value | %.2f |\n", costInfo.ThresholdValue))

	return sb.String()
}

// formatExpensiveOpsMarkdown formats expensive operations as markdown table
func formatExpensiveOpsMarkdown(ops []ExpensiveOperation) string {
	if len(ops) == 0 {
		return "_No expensive operations found_\n"
	}

	var sb strings.Builder

	sb.WriteString("| Operation | Cost | Details |\n")
	sb.WriteString("|-----------|------|---------||\n")

	for _, op := range ops {
		sb.WriteString(fmt.Sprintf("| %s | %.2f | %s |\n",
			escapeMarkdownSpecialChars(op.Operation),
			op.Cost,
			escapeMarkdownSpecialChars(op.Line)))
	}

	return sb.String()
}

// writeMarkdownPlan generates a Markdown file for analyze command
// Returns absolute path of generated file
func writeMarkdownPlan(plan, query, title string, costInfo *CostInfo) string {
	fileName := title + ".md"

	var sb strings.Builder

	// Title
	sb.WriteString("# Query Execution Plan\n\n")

	// Metadata
	sb.WriteString(fmt.Sprintf("**Generated:** %s  \n", time.Now().Format("January 2, 2006 15:04:05")))
	sb.WriteString(fmt.Sprintf("**Query:** %s\n\n", escapeMarkdownSpecialChars(query)))
	sb.WriteString("---\n\n")

	// Cost Analysis
	sb.WriteString("## Cost Analysis\n\n")
	sb.WriteString(formatCostInfoMarkdown(costInfo))
	sb.WriteString("\n")

	// Expensive Operations
	if costInfo != nil && len(costInfo.ExpensiveOps) > 0 {
		sb.WriteString("### Expensive Operations\n\n")
		sb.WriteString(formatExpensiveOpsMarkdown(costInfo.ExpensiveOps))
		sb.WriteString("\n")
	}

	sb.WriteString("---\n\n")

	// Execution Plan
	sb.WriteString("## Execution Plan\n\n")
	sb.WriteString("```\n")
	sb.WriteString(plan)
	sb.WriteString("\n```\n\n")

	sb.WriteString("---\n\n")
	sb.WriteString("**Note:** This plan was generated using PostgreSQL EXPLAIN\n")

	// Write to file
	file, err := os.Create(fileName)
	if err != nil {
		logErrorAndExit("unable to create Markdown file: ", err)
	}
	defer file.Close()

	_, err = file.WriteString(sb.String())
	if err != nil {
		logErrorAndExit("unable to write Markdown content: ", err)
	}

	abs, err := filepath.Abs(file.Name())
	if err != nil {
		logErrorAndExit("unable to get file absolute path: ", err)
	}

	return abs
}

// writeComparisonMarkdown generates a Markdown file for compare command
func writeComparisonMarkdown(result *ComparisonResult) {
	title := generateTitle()
	fileName := fmt.Sprintf("Comparison_%s.md", title)

	var sb strings.Builder

	// Title
	sb.WriteString("# Query Comparison Report\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format("January 2, 2006 15:04:05")))
	sb.WriteString("---\n\n")

	// Winner section
	winnerEmoji := "🏆"
	if result.Winner == "Tie" {
		winnerEmoji = "🤝"
	}
	sb.WriteString(fmt.Sprintf("## Winner: %s %s\n\n", result.Winner, winnerEmoji))

	// Comparison metrics table
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Cost Difference | %.2f |\n", result.CostDiff))
	sb.WriteString(fmt.Sprintf("| Percentage Difference | %.2f%% |\n", result.CostDiffPct))

	if result.CostDiff != 0 {
		var perfMultiplier string
		if result.CostDiff > 0 {
			mult := result.Cost1.TotalCost / result.Cost2.TotalCost
			perfMultiplier = fmt.Sprintf("Query 2 is %.2fx faster", mult)
		} else {
			mult := result.Cost2.TotalCost / result.Cost1.TotalCost
			perfMultiplier = fmt.Sprintf("Query 1 is %.2fx faster", mult)
		}
		sb.WriteString(fmt.Sprintf("| Performance Multiplier | %s |\n", perfMultiplier))
	}
	sb.WriteString("\n")

	// Recommendation
	sb.WriteString("**Recommendation:** ")
	sb.WriteString(escapeMarkdownSpecialChars(result.Recommendation))
	sb.WriteString("\n\n")
	sb.WriteString("---\n\n")

	// Query 1 section
	sb.WriteString("## Query 1\n\n")
	sb.WriteString("**SQL:**\n```sql\n")
	sb.WriteString(result.Query1)
	sb.WriteString("\n```\n\n")

	sb.WriteString("### Cost Analysis\n\n")
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Total Cost | %.2f |\n", result.Cost1.TotalCost))
	if len(result.Cost1.ExpensiveOps) > 0 {
		sb.WriteString(fmt.Sprintf("| Most Expensive Operation | %s (%.2f) |\n",
			escapeMarkdownSpecialChars(result.Cost1.ExpensiveOps[0].Operation),
			result.Cost1.ExpensiveOps[0].Cost))
	}
	sb.WriteString("\n")

	if len(result.Cost1.ExpensiveOps) > 0 {
		sb.WriteString("### Expensive Operations\n\n")
		sb.WriteString(formatExpensiveOpsMarkdown(result.Cost1.ExpensiveOps))
		sb.WriteString("\n")
	}

	sb.WriteString("### Execution Plan\n\n")
	sb.WriteString("```\n")
	sb.WriteString(result.Plan1)
	sb.WriteString("\n```\n\n")

	sb.WriteString("---\n\n")

	// Query 2 section
	sb.WriteString("## Query 2\n\n")
	sb.WriteString("**SQL:**\n```sql\n")
	sb.WriteString(result.Query2)
	sb.WriteString("\n```\n\n")

	sb.WriteString("### Cost Analysis\n\n")
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Total Cost | %.2f |\n", result.Cost2.TotalCost))
	if len(result.Cost2.ExpensiveOps) > 0 {
		sb.WriteString(fmt.Sprintf("| Most Expensive Operation | %s (%.2f) |\n",
			escapeMarkdownSpecialChars(result.Cost2.ExpensiveOps[0].Operation),
			result.Cost2.ExpensiveOps[0].Cost))
	}
	sb.WriteString("\n")

	if len(result.Cost2.ExpensiveOps) > 0 {
		sb.WriteString("### Expensive Operations\n\n")
		sb.WriteString(formatExpensiveOpsMarkdown(result.Cost2.ExpensiveOps))
		sb.WriteString("\n")
	}

	sb.WriteString("### Execution Plan\n\n")
	sb.WriteString("```\n")
	sb.WriteString(result.Plan2)
	sb.WriteString("\n```\n\n")

	sb.WriteString("---\n\n")

	// Detailed comparison table
	sb.WriteString("## Detailed Comparison\n\n")
	sb.WriteString("| Aspect | Query 1 | Query 2 |\n")
	sb.WriteString("|--------|---------|--------|\n")
	sb.WriteString(fmt.Sprintf("| Total Cost | %.2f | %.2f |\n", result.Cost1.TotalCost, result.Cost2.TotalCost))

	if len(result.Cost1.ExpensiveOps) > 0 || len(result.Cost2.ExpensiveOps) > 0 {
		topOp1 := "N/A"
		topOpCost1 := "N/A"
		if len(result.Cost1.ExpensiveOps) > 0 {
			topOp1 = escapeMarkdownSpecialChars(result.Cost1.ExpensiveOps[0].Operation)
			topOpCost1 = fmt.Sprintf("%.2f", result.Cost1.ExpensiveOps[0].Cost)
		}

		topOp2 := "N/A"
		topOpCost2 := "N/A"
		if len(result.Cost2.ExpensiveOps) > 0 {
			topOp2 = escapeMarkdownSpecialChars(result.Cost2.ExpensiveOps[0].Operation)
			topOpCost2 = fmt.Sprintf("%.2f", result.Cost2.ExpensiveOps[0].Cost)
		}

		sb.WriteString(fmt.Sprintf("| Top Operation | %s | %s |\n", topOp1, topOp2))
		sb.WriteString(fmt.Sprintf("| Top Op Cost | %s | %s |\n", topOpCost1, topOpCost2))
	}

	// Write to file
	file, err := os.Create(fileName)
	if err != nil {
		logErrorAndExit("unable to create Markdown file: ", err)
	}
	defer file.Close()

	_, err = file.WriteString(sb.String())
	if err != nil {
		logErrorAndExit("unable to write Markdown content: ", err)
	}

	abs, err := filepath.Abs(file.Name())
	if err != nil {
		logErrorAndExit("unable to get file absolute path: ", err)
	}

	// Display success message
	fmt.Println("\n" + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📁 Visual comparison report saved successfully!")
	fmt.Printf("   %s\n", abs)
	fmt.Printf("\n🏆 Winner: %s\n", result.Winner)
	if result.CostDiff != 0 {
		fmt.Printf("Cost Difference: %.2f (%.2f%%)\n", result.CostDiff, result.CostDiffPct)
	}
	fmt.Println("\n💡 Tip: Open this file in your markdown viewer to see the formatted comparison")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
}

// writeMarkdownBatchReport generates a Markdown file for batch command (combined mode)
// Returns absolute path of generated file
func writeMarkdownBatchReport(report BatchReport, fileName string) string {
	mdFileName := fileName + ".md"

	var sb strings.Builder

	// Title
	sb.WriteString("# Batch Analysis Report\n\n")
	sb.WriteString(fmt.Sprintf("**File:** %s  \n", report.FileName))
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", report.GeneratedAt.Format("January 2, 2006 15:04:05")))
	sb.WriteString("---\n\n")

	// Summary
	sb.WriteString("## Summary\n\n")
	sb.WriteString("| Metric | Count |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Total Queries | %d |\n", report.TotalQueries))
	sb.WriteString(fmt.Sprintf("| Successful | %d |\n", report.SuccessCount))
	sb.WriteString(fmt.Sprintf("| Failed | %d |\n", report.FailureCount))

	if report.TotalQueries > 0 {
		successRate := float64(report.SuccessCount) / float64(report.TotalQueries) * 100
		sb.WriteString(fmt.Sprintf("| Success Rate | %.1f%% |\n", successRate))
	}
	sb.WriteString("\n---\n\n")

	// Query Results
	sb.WriteString("## Query Results\n\n")

	for _, result := range report.Results {
		// Query header with status badge
		statusBadge := "✅"
		statusText := "Success"
		if result.Error != "" {
			statusBadge = "❌"
			statusText = "Failed"
		}
		sb.WriteString(fmt.Sprintf("### Query %d %s\n\n", result.QueryNumber, statusBadge))

		// SQL query
		sb.WriteString("**SQL:**\n```sql\n")
		sb.WriteString(result.Query)
		sb.WriteString("\n```\n\n")

		sb.WriteString(fmt.Sprintf("**Status:** %s  \n", statusText))

		// Handle errors
		if result.Error != "" {
			sb.WriteString(fmt.Sprintf("**Error:** %s\n\n", escapeMarkdownSpecialChars(result.Error)))
			sb.WriteString("---\n\n")
			continue
		}

		// Cost info
		if result.CostAnalysis != nil {
			sb.WriteString(fmt.Sprintf("**Total Cost:** %.2f\n\n", result.CostAnalysis.TotalCost))
		}

		// Execution Plan
		if result.ExecutionPlan != "" {
			sb.WriteString("#### Execution Plan\n```\n")
			sb.WriteString(result.ExecutionPlan)
			sb.WriteString("\n```\n\n")
		}

		// Cost Analysis
		if result.CostAnalysis != nil {
			sb.WriteString("#### Cost Analysis\n\n")
			sb.WriteString(formatCostInfoMarkdown(result.CostAnalysis))
			sb.WriteString("\n")

			if len(result.CostAnalysis.ExpensiveOps) > 0 {
				sb.WriteString("##### Expensive Operations\n\n")
				sb.WriteString(formatExpensiveOpsMarkdown(result.CostAnalysis.ExpensiveOps))
				sb.WriteString("\n")
			}
		}

		sb.WriteString("---\n\n")
	}

	// Performance Summary (if applicable)
	if report.SuccessCount > 0 {
		sb.WriteString("## Performance Summary\n\n")
		sb.WriteString("| Query # | Cost | Status |\n")
		sb.WriteString("|---------|------|--------|\n")

		for _, result := range report.Results {
			statusEmoji := "✅ Success"
			costStr := "N/A"

			if result.Error != "" {
				statusEmoji = "❌ Failed"
			} else if result.CostAnalysis != nil {
				costStr = fmt.Sprintf("%.2f", result.CostAnalysis.TotalCost)
			}

			sb.WriteString(fmt.Sprintf("| %d | %s | %s |\n", result.QueryNumber, costStr, statusEmoji))
		}
	}

	// Write to file
	file, err := os.Create(mdFileName)
	if err != nil {
		logErrorAndExit("unable to create Markdown file: ", err)
	}
	defer file.Close()

	_, err = file.WriteString(sb.String())
	if err != nil {
		logErrorAndExit("unable to write Markdown content: ", err)
	}

	abs, err := filepath.Abs(file.Name())
	if err != nil {
		logErrorAndExit("unable to get file absolute path: ", err)
	}

	return abs
}
