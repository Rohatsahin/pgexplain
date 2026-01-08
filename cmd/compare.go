/*
Package cmd

# Copyright © 2024 Rohat Sahin

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare execution plans of two SQL queries",
	Long:  "Generate and compare execution plans for two queries side-by-side to identify performance differences",
	Args:  cobra.ExactArgs(2),
	Run:   runCompare,
}

type ComparisonResult struct {
	Query1        string     `json:"query1"`
	Query2        string     `json:"query2"`
	Plan1         string     `json:"plan1"`
	Plan2         string     `json:"plan2"`
	Cost1         *CostInfo  `json:"cost_analysis1"`
	Cost2         *CostInfo  `json:"cost_analysis2"`
	Winner        string     `json:"winner"`
	CostDiff      float64    `json:"cost_difference"`
	CostDiffPct   float64    `json:"cost_difference_percentage"`
	Recommendation string    `json:"recommendation"`
}

func runCompare(cmd *cobra.Command, args []string) {
	query1 := args[0]
	query2 := args[1]

	// Load configuration
	config, _ := loadConfig()

	fmt.Println("\n🔬 Starting query comparison...")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	fmt.Println("🔍 Analyzing Query 1...")
	plan1, err := generateExecutionPlan(query1, config)
	if err != nil {
		fmt.Println("❌ Failed to analyze Query 1")
		logErrorAndExit("Error: ", err)
	}
	fmt.Println("✅ Query 1 complete!")

	fmt.Println("\n🔍 Analyzing Query 2...")
	plan2, err := generateExecutionPlan(query2, config)
	if err != nil {
		fmt.Println("❌ Failed to analyze Query 2")
		logErrorAndExit("Error: ", err)
	}
	fmt.Println("✅ Query 2 complete!")
	fmt.Println()

	// Parse costs for both queries
	cost1 := parseCost(plan1, 0)
	cost2 := parseCost(plan2, 0)

	// Create comparison result
	result := &ComparisonResult{
		Query1:   query1,
		Query2:   query2,
		Plan1:    plan1,
		Plan2:    plan2,
		Cost1:    cost1,
		Cost2:    cost2,
		CostDiff: cost1.TotalCost - cost2.TotalCost,
	}

	// Calculate percentage difference
	if cost2.TotalCost != 0 {
		result.CostDiffPct = (result.CostDiff / cost2.TotalCost) * 100
	}

	// Determine winner
	if cost1.TotalCost < cost2.TotalCost {
		result.Winner = "Query 1"
		result.Recommendation = "Query 1 is more efficient. Consider using this approach."
	} else if cost2.TotalCost < cost1.TotalCost {
		result.Winner = "Query 2"
		result.Recommendation = "Query 2 is more efficient. Consider using this approach."
	} else {
		result.Winner = "Tie"
		result.Recommendation = "Both queries have similar costs. Choose based on readability and maintainability."
	}

	// Output format
	format, _ := cmd.Flags().GetString("format")

	switch format {
	case "json":
		writeComparisonJSON(result)
	case "text":
		displayComparisonText(result)
	default:
		logErrorAndExit("Invalid format specified", fmt.Errorf("supported formats: text, json"))
	}
}

func displayComparisonText(result *ComparisonResult) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("QUERY COMPARISON REPORT")
	fmt.Println(strings.Repeat("=", 80))

	// Query 1
	fmt.Println("\nQuery 1:")
	fmt.Printf("  %s\n", result.Query1)
	fmt.Printf("  Total Cost: %.2f\n", result.Cost1.TotalCost)
	if len(result.Cost1.ExpensiveOps) > 0 {
		fmt.Printf("  Most Expensive Operation: %s (%.2f)\n",
			result.Cost1.ExpensiveOps[0].Operation,
			result.Cost1.ExpensiveOps[0].Cost)
	}

	fmt.Println(strings.Repeat("-", 80))

	// Query 2
	fmt.Println("\nQuery 2:")
	fmt.Printf("  %s\n", result.Query2)
	fmt.Printf("  Total Cost: %.2f\n", result.Cost2.TotalCost)
	if len(result.Cost2.ExpensiveOps) > 0 {
		fmt.Printf("  Most Expensive Operation: %s (%.2f)\n",
			result.Cost2.ExpensiveOps[0].Operation,
			result.Cost2.ExpensiveOps[0].Cost)
	}

	fmt.Println(strings.Repeat("=", 80))

	// Comparison
	fmt.Println("\nCOMPARISON RESULTS")
	fmt.Println(strings.Repeat("-", 80))

	// Add winner emoji
	winnerEmoji := "🏆"
	if result.Winner == "Tie" {
		winnerEmoji = "🤝"
	}
	fmt.Printf("Winner: %s %s\n", winnerEmoji, result.Winner)
	fmt.Printf("Cost Difference: %.2f (%.2f%%)\n", result.CostDiff, result.CostDiffPct)

	if result.CostDiff != 0 {
		if result.CostDiff > 0 {
			fmt.Printf("⚡ Query 2 is %.2fx faster\n", (result.Cost1.TotalCost / result.Cost2.TotalCost))
		} else {
			fmt.Printf("⚡ Query 1 is %.2fx faster\n", (result.Cost2.TotalCost / result.Cost1.TotalCost))
		}
	}

	fmt.Printf("\n💡 Recommendation: %s\n", result.Recommendation)
	fmt.Println(strings.Repeat("=", 80))

	// Detailed Plans
	fmt.Println("\nDETAILED EXECUTION PLANS")
	fmt.Println(strings.Repeat("-", 80))

	fmt.Println("\n[Query 1 Execution Plan]")
	fmt.Println(result.Plan1)

	fmt.Println("\n" + strings.Repeat("-", 80))
	fmt.Println("\n[Query 2 Execution Plan]")
	fmt.Println(result.Plan2)

	fmt.Println(strings.Repeat("=", 80) + "\n")
}

func writeComparisonJSON(result *ComparisonResult) {
	fmt.Println("💾 Saving comparison as JSON...")
	title := generateTitle()
	fileName := fmt.Sprintf("Comparison_%s.json", title)

	writeJSONToFile(fileName, result)

	fmt.Println("\n" + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📁 Comparison saved successfully!")
	fmt.Printf("   %s\n", fileName)

	// Show quick summary
	winnerEmoji := "🏆"
	if result.Winner == "Tie" {
		winnerEmoji = "🤝"
	}
	fmt.Printf("\n%s Winner: %s (Cost diff: %.2f%%)\n", winnerEmoji, result.Winner, result.CostDiffPct)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
}

func init() {
	compareCmd.Flags().StringP("format", "f", "text", "Output format (text or json)")
	rootCmd.AddCommand(compareCmd)
}