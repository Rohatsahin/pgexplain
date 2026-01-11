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
	"bufio"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var compareCmd = &cobra.Command{
	Use:   "compare [QUERY1] [QUERY2]",
	Short: "Compare execution plans of two SQL queries",
	Long:  "Generate and compare execution plans for two queries side-by-side to identify performance differences",
	Args:  cobra.MaximumNArgs(2),
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
	// Get queries from file flags or arguments
	query1, query2, err := getCompareQueryInput(cmd, args)
	if err != nil {
		logErrorAndExit("Failed to get query input: ", err)
	}

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
	case "html":
		writeComparisonHTML(result)
	case "markdown":
		writeComparisonMarkdown(result)
	case "csv":
		writeComparisonCSV(result)
	default:
		logErrorAndExit("Invalid format specified", fmt.Errorf("supported formats: text, json, html, markdown, csv"))
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

func writeComparisonHTML(result *ComparisonResult) {
	fmt.Println("💾 Generating visual comparison report...")

	title := generateTitle()
	fileName := fmt.Sprintf("Comparison_%s.html", title)

	// Determine winner styling
	winnerEmoji := "🏆"
	winnerClass := "winner"
	if result.Winner == "Tie" {
		winnerEmoji = "🤝"
		winnerClass = "tie"
	}

	// Calculate performance multiplier
	var perfMultiplier string
	if result.CostDiff != 0 {
		if result.CostDiff > 0 {
			mult := result.Cost1.TotalCost / result.Cost2.TotalCost
			perfMultiplier = fmt.Sprintf("Query 2 is %.2fx faster", mult)
		} else {
			mult := result.Cost2.TotalCost / result.Cost1.TotalCost
			perfMultiplier = fmt.Sprintf("Query 1 is %.2fx faster", mult)
		}
	} else {
		perfMultiplier = "Both queries have identical cost"
	}

	// Calculate cost bar widths for visualization
	maxCost := math.Max(result.Cost1.TotalCost, result.Cost2.TotalCost)
	cost1Width := 100.0
	cost2Width := 100.0
	if maxCost > 0 {
		cost1Width = (result.Cost1.TotalCost / maxCost) * 100
		cost2Width = (result.Cost2.TotalCost / maxCost) * 100
	}

	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Query Comparison - Visual Diff</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.2/dist/css/bootstrap.min.css" rel="stylesheet">
    <style>
        body {
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            min-height: 100vh;
        }
        .container {
            max-width: 1600px;
            background-color: white;
            padding: 40px;
            border-radius: 12px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
        }
        .header {
            text-align: center;
            margin-bottom: 40px;
            padding-bottom: 20px;
            border-bottom: 3px solid #667eea;
        }
        .header h1 { color: #667eea; font-weight: bold; }

        /* Winner Badge */
        .winner-badge {
            display: inline-block;
            padding: 15px 30px;
            border-radius: 50px;
            font-size: 1.5em;
            font-weight: bold;
            margin: 20px 0;
            box-shadow: 0 4px 15px rgba(0,0,0,0.1);
        }
        .winner-badge.winner {
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
        }
        .winner-badge.tie {
            background: linear-gradient(135deg, #f093fb 0%%, #f5576c 100%%);
            color: white;
        }

        /* Cost Comparison */
        .cost-comparison {
            background: #f8f9fa;
            padding: 30px;
            border-radius: 12px;
            margin: 30px 0;
        }
        .cost-bar-container {
            margin: 20px 0;
        }
        .cost-bar {
            height: 50px;
            border-radius: 8px;
            display: flex;
            align-items: center;
            padding: 0 15px;
            color: white;
            font-weight: bold;
            transition: all 0.3s ease;
            margin-bottom: 15px;
        }
        .cost-bar:hover { transform: translateX(5px); }
        .cost-bar-1 { background: linear-gradient(90deg, #667eea 0%%, #764ba2 100%%); }
        .cost-bar-2 { background: linear-gradient(90deg, #f093fb 0%%, #f5576c 100%%); }

        /* Side by Side Comparison */
        .comparison-grid {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 20px;
            margin-top: 30px;
        }
        .query-panel {
            background: #f8f9fa;
            padding: 25px;
            border-radius: 12px;
            border: 2px solid #dee2e6;
            transition: all 0.3s ease;
        }
        .query-panel:hover {
            border-color: #667eea;
            box-shadow: 0 5px 20px rgba(102, 126, 234, 0.2);
        }
        .query-panel h3 {
            color: #667eea;
            margin-bottom: 20px;
            font-weight: bold;
        }
        .query-sql {
            background: #ffffff;
            padding: 20px;
            border-radius: 8px;
            font-family: 'Courier New', monospace;
            font-size: 0.9em;
            margin-bottom: 20px;
            border-left: 4px solid #667eea;
            overflow-x: auto;
        }
        .execution-plan {
            background: #ffffff;
            padding: 20px;
            border-radius: 8px;
            font-family: 'Courier New', monospace;
            white-space: pre-wrap;
            font-size: 0.85em;
            max-height: 600px;
            overflow-y: auto;
            border: 1px solid #dee2e6;
        }

        /* Stats Cards */
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin: 30px 0;
        }
        .stat-card {
            background: white;
            padding: 20px;
            border-radius: 12px;
            text-align: center;
            border: 2px solid #dee2e6;
            transition: all 0.3s ease;
        }
        .stat-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 5px 20px rgba(0,0,0,0.1);
        }
        .stat-value {
            font-size: 2em;
            font-weight: bold;
            color: #667eea;
            margin: 10px 0;
        }
        .stat-label {
            color: #6c757d;
            font-size: 0.9em;
            text-transform: uppercase;
            letter-spacing: 1px;
        }

        /* Recommendation */
        .recommendation {
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
            padding: 25px;
            border-radius: 12px;
            margin: 30px 0;
            text-align: center;
            font-size: 1.1em;
        }

        /* Expensive Operations */
        .expensive-ops {
            margin-top: 15px;
        }
        .expensive-ops h5 {
            color: #6c757d;
            font-size: 0.9em;
            margin-bottom: 10px;
        }
        .op-badge {
            display: inline-block;
            background: #ffc107;
            color: #000;
            padding: 5px 12px;
            border-radius: 20px;
            margin: 5px;
            font-size: 0.85em;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔬 Query Comparison Report</h1>
            <p class="text-muted">Visual execution plan diff</p>
            <div class="winner-badge %s">
                %s %s
            </div>
        </div>

        <!-- Cost Comparison -->
        <div class="cost-comparison">
            <h2 class="text-center mb-4">📊 Cost Analysis</h2>
            <div class="cost-bar-container">
                <div class="cost-bar cost-bar-1" style="width: %.2f%%%%">
                    Query 1: %.2f
                </div>
                <div class="cost-bar cost-bar-2" style="width: %.2f%%%%">
                    Query 2: %.2f
                </div>
            </div>

            <div class="stats-grid">
                <div class="stat-card">
                    <div class="stat-label">Cost Difference</div>
                    <div class="stat-value">%.2f</div>
                </div>
                <div class="stat-card">
                    <div class="stat-label">Percentage</div>
                    <div class="stat-value">%.2f%%%%</div>
                </div>
                <div class="stat-card">
                    <div class="stat-label">Performance</div>
                    <div class="stat-value" style="font-size: 1.2em;">%s</div>
                </div>
            </div>
        </div>

        <!-- Recommendation -->
        <div class="recommendation">
            💡 <strong>Recommendation:</strong> %s
        </div>

        <!-- Side by Side Comparison -->
        <div class="comparison-grid">
            <!-- Query 1 -->
            <div class="query-panel">
                <h3>Query 1</h3>
                <div class="query-sql">%s</div>
                <div class="stat-card">
                    <div class="stat-label">Total Cost</div>
                    <div class="stat-value">%.2f</div>
                </div>`,
		winnerClass, winnerEmoji, result.Winner,
		cost1Width, result.Cost1.TotalCost,
		cost2Width, result.Cost2.TotalCost,
		math.Abs(result.CostDiff),
		math.Abs(result.CostDiffPct),
		perfMultiplier,
		result.Recommendation,
		result.Query1,
		result.Cost1.TotalCost)

	// Add expensive operations for Query 1
	if len(result.Cost1.ExpensiveOps) > 0 {
		htmlContent += `
                <div class="expensive-ops">
                    <h5>Expensive Operations:</h5>`
		for _, op := range result.Cost1.ExpensiveOps {
			if len(result.Cost1.ExpensiveOps) <= 3 {
				htmlContent += fmt.Sprintf(`
                    <span class="op-badge">%s (%.2f)</span>`, op.Operation, op.Cost)
			}
		}
		htmlContent += `
                </div>`
	}

	htmlContent += fmt.Sprintf(`
                <h4 class="mt-4 mb-3">Execution Plan:</h4>
                <div class="execution-plan">%s</div>
            </div>

            <!-- Query 2 -->
            <div class="query-panel">
                <h3>Query 2</h3>
                <div class="query-sql">%s</div>
                <div class="stat-card">
                    <div class="stat-label">Total Cost</div>
                    <div class="stat-value">%.2f</div>
                </div>`,
		result.Plan1,
		result.Query2,
		result.Cost2.TotalCost)

	// Add expensive operations for Query 2
	if len(result.Cost2.ExpensiveOps) > 0 {
		htmlContent += `
                <div class="expensive-ops">
                    <h5>Expensive Operations:</h5>`
		for _, op := range result.Cost2.ExpensiveOps {
			if len(result.Cost2.ExpensiveOps) <= 3 {
				htmlContent += fmt.Sprintf(`
                    <span class="op-badge">%s (%.2f)</span>`, op.Operation, op.Cost)
			}
		}
		htmlContent += `
                </div>`
	}

	htmlContent += fmt.Sprintf(`
                <h4 class="mt-4 mb-3">Execution Plan:</h4>
                <div class="execution-plan">%s</div>
            </div>
        </div>
    </div>
</body>
</html>`, result.Plan2)

	file, err := os.Create(fileName)
	if err != nil {
		logErrorAndExit("unable to create comparison HTML file: ", err)
	}
	defer file.Close()

	_, err = file.WriteString(htmlContent)
	if err != nil {
		logErrorAndExit("unable to write comparison HTML file: ", err)
	}

	abs, err := filepath.Abs(file.Name())
	if err != nil {
		logErrorAndExit("unable to get comparison file absolute path: ", err)
	}

	fmt.Println("\n" + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📁 Visual comparison report saved successfully!")
	fmt.Printf("   %s\n", abs)

	// Show quick summary
	fmt.Printf("\n%s Winner: %s\n", winnerEmoji, result.Winner)
	fmt.Printf("Cost Difference: %.2f (%.2f%%)\n", math.Abs(result.CostDiff), math.Abs(result.CostDiffPct))
	fmt.Println("\n💡 Tip: Open this file in your browser to view the interactive visual diff")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
}

// getCompareQueryInput retrieves two SQL queries from various input sources
// Priority: --file1/--file2 flags > command arguments > --editor flag > interactive prompts
func getCompareQueryInput(cmd *cobra.Command, args []string) (string, string, error) {
	file1, _ := cmd.Flags().GetString("file1")
	file2, _ := cmd.Flags().GetString("file2")
	useEditor, _ := cmd.Flags().GetBool("editor")

	var query1, query2 string
	var err error

	// Get first query
	if file1 != "" {
		content, err := os.ReadFile(file1)
		if err != nil {
			return "", "", fmt.Errorf("failed to read file1 %s: %w", file1, err)
		}
		query1 = strings.TrimSpace(string(content))
		if query1 == "" {
			return "", "", fmt.Errorf("file1 %s is empty", file1)
		}
	} else if len(args) > 0 {
		query1 = args[0]
	} else if useEditor {
		fmt.Println("\n📝 Query 1:")
		query1, err = getQueryFromEditorCompare("Query 1")
		if err != nil {
			return "", "", err
		}
	} else {
		fmt.Println("\n📝 Enter Query 1 (paste or type, press Ctrl+D when done):")
		query1, err = getQueryFromPromptCompare()
		if err != nil {
			return "", "", fmt.Errorf("failed to get query 1: %w", err)
		}
	}

	// Get second query
	if file2 != "" {
		content, err := os.ReadFile(file2)
		if err != nil {
			return "", "", fmt.Errorf("failed to read file2 %s: %w", file2, err)
		}
		query2 = strings.TrimSpace(string(content))
		if query2 == "" {
			return "", "", fmt.Errorf("file2 %s is empty", file2)
		}
	} else if len(args) > 1 {
		query2 = args[1]
	} else if useEditor {
		fmt.Println("\n📝 Query 2:")
		query2, err = getQueryFromEditorCompare("Query 2")
		if err != nil {
			return "", "", err
		}
	} else {
		fmt.Println("\n📝 Enter Query 2 (paste or type, press Ctrl+D when done):")
		query2, err = getQueryFromPromptCompare()
		if err != nil {
			return "", "", fmt.Errorf("failed to get query 2: %w", err)
		}
	}

	return query1, query2, err
}

// getQueryFromEditorCompare opens editor for compare command
func getQueryFromEditorCompare(queryName string) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vim"
	}

	tmpFile, err := os.CreateTemp("", "pgexplain_compare_*.sql")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	placeholder := fmt.Sprintf("-- Enter %s below, then save and close the editor\n\n", queryName)
	if err := os.WriteFile(tmpPath, []byte(placeholder), 0644); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	fmt.Printf("✏️  Opening editor for %s: %s\n", queryName, editor)
	fmt.Println("💡 Write your query, save, and close the editor to continue...")

	cmd := exec.Command(editor, tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor exited with error: %w", err)
	}

	content, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", fmt.Errorf("failed to read temp file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var queryLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "--") && line != "" {
			queryLines = append(queryLines, line)
		}
	}
	query := strings.TrimSpace(strings.Join(queryLines, "\n"))

	if query == "" {
		return "", fmt.Errorf("no query entered in editor")
	}

	return query, nil
}

// getQueryFromPromptCompare prompts for query input in compare command
func getQueryFromPromptCompare() (string, error) {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	var lines []string
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	query := strings.TrimSpace(strings.Join(lines, "\n"))
	if query == "" {
		return "", fmt.Errorf("no query entered")
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ Query received!")

	return query, nil
}

func init() {
	compareCmd.Flags().StringP("format", "f", "text", "Output format (text, json, html, markdown, or csv)")
	compareCmd.Flags().StringP("file1", "", "", "Read first SQL query from file")
	compareCmd.Flags().StringP("file2", "", "", "Read second SQL query from file")
	compareCmd.Flags().BoolP("editor", "e", false, "Open $EDITOR to write/paste queries")
	rootCmd.AddCommand(compareCmd)
}