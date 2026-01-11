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
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var batchCmd = &cobra.Command{
	Use:   "batch [SQL_FILE]",
	Short: "Analyze multiple SQL queries from a file",
	Long: `Batch analyze SQL queries from a file. Queries should be separated by semicolons (;).
Empty lines and SQL comments (--) are automatically ignored.

Example:
  pg_explain batch queries.sql
  pg_explain batch queries.sql --format json
  pg_explain batch queries.sql --combined --output-dir ./reports`,
	Args: cobra.ExactArgs(1),
	Run:  runBatch,
}

// BatchResult stores the analysis result for a single query
type BatchResult struct {
	QueryNumber   int        `json:"query_number"`
	Query         string     `json:"query"`
	ExecutionPlan string     `json:"execution_plan"`
	CostAnalysis  *CostInfo  `json:"cost_analysis,omitempty"`
	Error         string     `json:"error,omitempty"`
	GeneratedAt   time.Time  `json:"generated_at"`
}

// BatchReport stores all batch analysis results
type BatchReport struct {
	FileName      string        `json:"file_name"`
	TotalQueries  int           `json:"total_queries"`
	SuccessCount  int           `json:"success_count"`
	FailureCount  int           `json:"failure_count"`
	Results       []BatchResult `json:"results"`
	GeneratedAt   time.Time     `json:"generated_at"`
}

func runBatch(cmd *cobra.Command, args []string) {
	sqlFile := args[0]

	// Load configuration
	config, _ := loadConfig()

	// Get flag values
	threshold, _ := cmd.Flags().GetFloat64("threshold")
	if !cmd.Flags().Changed("threshold") && config.Defaults.Threshold > 0 {
		threshold = config.Defaults.Threshold
	}

	format, _ := cmd.Flags().GetString("format")
	if !cmd.Flags().Changed("format") && config.Defaults.Format != "" {
		format = config.Defaults.Format
	}

	combined, _ := cmd.Flags().GetBool("combined")
	outputDir, _ := cmd.Flags().GetString("output-dir")
	recommendIndexes, _ := cmd.Flags().GetBool("recommend-indexes")
	indexThreshold, _ := cmd.Flags().GetFloat64("index-threshold")
	continueOnError, _ := cmd.Flags().GetBool("continue-on-error")

	// Show friendly start message
	fmt.Println("\n🔍 Starting batch analysis...")
	fmt.Printf("📁 SQL file: %s\n", sqlFile)
	fmt.Printf("📊 Output format: %s\n", format)
	if threshold > 0 {
		fmt.Printf("⚡ Cost threshold: %.0f\n", threshold)
	}
	if combined {
		fmt.Println("📦 Mode: Combined report")
	} else {
		fmt.Println("📦 Mode: Individual files")
	}
	fmt.Println()

	// Read and parse SQL file
	queries, err := parseSQLFile(sqlFile)
	if err != nil {
		fmt.Println("❌ Failed to read SQL file")
		logErrorAndExit("Error: ", err)
	}

	if len(queries) == 0 {
		fmt.Println("⚠️  No valid SQL queries found in file")
		return
	}

	fmt.Printf("✅ Found %d queries to analyze\n\n", len(queries))

	// Create output directory if specified
	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			logErrorAndExit("Failed to create output directory: ", err)
		}
	}

	// Process queries
	batchReport := BatchReport{
		FileName:    filepath.Base(sqlFile),
		GeneratedAt: time.Now(),
		Results:     make([]BatchResult, 0),
	}

	for i, query := range queries {
		queryNum := i + 1
		fmt.Printf("🔄 Processing query %d/%d...\n", queryNum, len(queries))

		result := BatchResult{
			QueryNumber: queryNum,
			Query:       query,
			GeneratedAt: time.Now(),
		}

		plan, err := generateExecutionPlan(query, config)
		if err != nil {
			result.Error = err.Error()
			batchReport.FailureCount++
			fmt.Printf("   ❌ Query %d failed: %v\n\n", queryNum, err)

			if !continueOnError {
				fmt.Println("⛔ Stopping batch analysis due to error. Use --continue-on-error to skip failed queries.")
				break
			}
		} else {
			result.ExecutionPlan = plan
			batchReport.SuccessCount++

			// Cost analysis
			if threshold > 0 {
				costInfo := parseCost(plan, threshold)
				result.CostAnalysis = costInfo
				if costInfo.ExceedsLimit {
					fmt.Printf("   ⚠️  Query %d exceeds cost threshold (%.2f > %.0f)\n", queryNum, costInfo.TotalCost, threshold)
				} else {
					fmt.Printf("   ✅ Query %d cost: %.2f\n", queryNum, costInfo.TotalCost)
				}
			} else {
				fmt.Printf("   ✅ Query %d analyzed successfully\n", queryNum)
			}

			// Index recommendations
			if recommendIndexes {
				indexInfo := analyzeIndexOpportunities(plan, indexThreshold)
				if indexInfo.TotalFound > 0 {
					fmt.Printf("   💡 Found %d index recommendations\n", indexInfo.TotalFound)
				}
			}
			fmt.Println()
		}

		batchReport.Results = append(batchReport.Results, result)
	}

	batchReport.TotalQueries = len(queries)

	// Generate output
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📊 Batch Analysis Complete\n")
	fmt.Printf("   Total: %d | Success: %d | Failed: %d\n",
		batchReport.TotalQueries, batchReport.SuccessCount, batchReport.FailureCount)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	if combined {
		// Generate combined report
		fmt.Println("💾 Generating combined report...")
		fileName := generateBatchFileName(sqlFile, format, outputDir)

		switch format {
		case "json":
			absPath := writeJSONToFile(fileName, batchReport)
			fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Println("📁 Batch report saved successfully!")
			fmt.Printf("   %s\n", absPath)
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		case "html":
			absPath := writeBatchHTMLReport(batchReport, fileName)
			fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Println("📁 Batch report saved successfully!")
			fmt.Printf("   %s\n", absPath)
			fmt.Println("\n💡 Tip: Open this file in your browser to view all query plans")
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		case "markdown":
			absPath := writeMarkdownBatchReport(batchReport, fileName)
			fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Println("📁 Batch report saved successfully!")
			fmt.Printf("   %s\n", absPath)
			fmt.Println("\n💡 Tip: Open this file in your markdown viewer to view all query plans")
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		case "csv":
			absPath := writeCSVBatchReport(batchReport, fileName)
			fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Println("📁 Batch report saved successfully!")
			fmt.Printf("   %s\n", absPath)
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		default:
			logErrorAndExit("Invalid format specified", fmt.Errorf("supported formats: html, json, markdown, csv"))
		}
	} else {
		// Generate individual files
		fmt.Println("💾 Generating individual files...")
		savedFiles := make([]string, 0)

		for _, result := range batchReport.Results {
			if result.Error != "" {
				continue // Skip failed queries
			}

			title := fmt.Sprintf("Query_%d_%s", result.QueryNumber, generateTitle())
			fileName := filepath.Join(outputDir, title)

			switch format {
			case "json":
				absPath := writeJSONPlan(result.ExecutionPlan, result.Query, fileName, result.CostAnalysis)
				savedFiles = append(savedFiles, absPath)
			case "html":
				absPath := writePlan(result.ExecutionPlan, result.Query, fileName)
				savedFiles = append(savedFiles, absPath)
			case "markdown":
				absPath := writeMarkdownPlan(result.ExecutionPlan, result.Query, fileName, result.CostAnalysis)
				savedFiles = append(savedFiles, absPath)
			case "csv":
				absPath := writeCSVPlan(result.ExecutionPlan, result.Query, fileName, result.CostAnalysis)
				savedFiles = append(savedFiles, absPath)
			}
		}

		fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("📁 Generated %d files successfully!\n", len(savedFiles))
		if len(savedFiles) > 0 && len(savedFiles) <= 5 {
			for _, file := range savedFiles {
				fmt.Printf("   %s\n", file)
			}
		} else if len(savedFiles) > 5 {
			fmt.Printf("   %s\n", savedFiles[0])
			fmt.Printf("   ... and %d more files\n", len(savedFiles)-1)
		}
		if outputDir != "" {
			fmt.Printf("\n💡 All files saved to: %s\n", outputDir)
		}
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	}
}

// parseSQLFile reads a SQL file and extracts individual queries
// Queries are separated by semicolons, comments and empty lines are ignored
func parseSQLFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open SQL file: %w", err)
	}
	defer file.Close()

	var queries []string
	var currentQuery strings.Builder
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "--") || strings.HasPrefix(line, "/*") {
			continue
		}

		// Add line to current query
		currentQuery.WriteString(line)
		currentQuery.WriteString(" ")

		// Check if query ends with semicolon
		if strings.HasSuffix(line, ";") {
			query := strings.TrimSpace(currentQuery.String())
			// Remove trailing semicolon for EXPLAIN
			query = strings.TrimSuffix(query, ";")

			if query != "" {
				queries = append(queries, query)
			}
			currentQuery.Reset()
		}
	}

	// Handle last query if it doesn't end with semicolon
	if currentQuery.Len() > 0 {
		query := strings.TrimSpace(currentQuery.String())
		query = strings.TrimSuffix(query, ";")
		if query != "" {
			queries = append(queries, query)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading SQL file: %w", err)
	}

	return queries, nil
}

// generateBatchFileName creates a filename for the batch report
func generateBatchFileName(sqlFile, format, outputDir string) string {
	baseName := strings.TrimSuffix(filepath.Base(sqlFile), filepath.Ext(sqlFile))
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	fileName := fmt.Sprintf("Batch_%s_%s.%s", baseName, timestamp, format)

	if outputDir != "" {
		return filepath.Join(outputDir, fileName)
	}
	return fileName
}

// writeBatchHTMLReport generates an HTML report for batch analysis
func writeBatchHTMLReport(report BatchReport, fileName string) string {
	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Batch Analysis Report - %s</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.2/dist/css/bootstrap.min.css" rel="stylesheet">
    <style>
        body { padding: 20px; background-color: #f5f5f5; }
        .container { max-width: 1400px; background-color: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { margin-bottom: 30px; }
        .stats { display: flex; gap: 20px; margin-bottom: 30px; }
        .stat-card { flex: 1; padding: 20px; border-radius: 8px; text-align: center; }
        .stat-card.total { background-color: #e3f2fd; }
        .stat-card.success { background-color: #e8f5e9; }
        .stat-card.failed { background-color: #ffebee; }
        .stat-number { font-size: 2em; font-weight: bold; margin-bottom: 5px; }
        .query-card { margin-bottom: 20px; border: 1px solid #ddd; border-radius: 8px; overflow: hidden; }
        .query-header { background-color: #f8f9fa; padding: 15px; border-bottom: 1px solid #ddd; cursor: pointer; }
        .query-header:hover { background-color: #e9ecef; }
        .query-body { padding: 15px; display: none; }
        .query-body.show { display: block; }
        .query-sql { background-color: #f5f5f5; padding: 15px; border-radius: 4px; font-family: monospace; margin-bottom: 15px; }
        .execution-plan { background-color: #f8f9fa; padding: 15px; border-radius: 4px; font-family: monospace; white-space: pre-wrap; font-size: 0.9em; }
        .badge { margin-left: 10px; }
        .cost-info { margin-top: 15px; padding: 10px; background-color: #fff3cd; border-radius: 4px; }
        .error-info { margin-top: 15px; padding: 10px; background-color: #f8d7da; border-radius: 4px; color: #721c24; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📊 Batch Analysis Report</h1>
            <p class="text-muted">File: %s | Generated: %s</p>
        </div>

        <div class="stats">
            <div class="stat-card total">
                <div class="stat-number">%d</div>
                <div>Total Queries</div>
            </div>
            <div class="stat-card success">
                <div class="stat-number">%d</div>
                <div>Successful</div>
            </div>
            <div class="stat-card failed">
                <div class="stat-number">%d</div>
                <div>Failed</div>
            </div>
        </div>

        <div class="queries">`,
		report.FileName,
		report.FileName,
		report.GeneratedAt.Format("January 2, 2006 15:04:05"),
		report.TotalQueries,
		report.SuccessCount,
		report.FailureCount)

	// Add each query
	for _, result := range report.Results {
		statusBadge := `<span class="badge bg-success">Success</span>`
		if result.Error != "" {
			statusBadge = `<span class="badge bg-danger">Failed</span>`
		}

		htmlContent += fmt.Sprintf(`
            <div class="query-card">
                <div class="query-header" onclick="toggleQuery(%d)">
                    <strong>Query %d</strong>
                    %s
                    <span class="float-end">▼</span>
                </div>
                <div class="query-body" id="query-%d">
                    <h5>SQL Query:</h5>
                    <div class="query-sql">%s</div>`,
			result.QueryNumber,
			result.QueryNumber,
			statusBadge,
			result.QueryNumber,
			result.Query)

		if result.Error != "" {
			htmlContent += fmt.Sprintf(`
                    <div class="error-info">
                        <strong>Error:</strong> %s
                    </div>`, result.Error)
		} else {
			htmlContent += fmt.Sprintf(`
                    <h5>Execution Plan:</h5>
                    <div class="execution-plan">%s</div>`, result.ExecutionPlan)

			if result.CostAnalysis != nil {
				htmlContent += fmt.Sprintf(`
                    <div class="cost-info">
                        <strong>Cost Analysis:</strong> Total Cost: %.2f`, result.CostAnalysis.TotalCost)

				if result.CostAnalysis.ExceedsLimit {
					htmlContent += fmt.Sprintf(` | ⚠️ Exceeds threshold (%.0f)`, result.CostAnalysis.ThresholdValue)
				}

				htmlContent += `</div>`
			}
		}

		htmlContent += `
                </div>
            </div>`
	}

	htmlContent += `
        </div>
    </div>

    <script>
        function toggleQuery(queryNum) {
            const queryBody = document.getElementById('query-' + queryNum);
            queryBody.classList.toggle('show');
        }

        // Expand first query by default
        if (document.getElementById('query-1')) {
            document.getElementById('query-1').classList.add('show');
        }
    </script>
</body>
</html>`

	file, err := os.Create(fileName)
	if err != nil {
		logErrorAndExit("unable to create batch HTML report: ", err)
	}
	defer file.Close()

	_, err = file.WriteString(htmlContent)
	if err != nil {
		logErrorAndExit("unable to write batch HTML report: ", err)
	}

	abs, err := filepath.Abs(file.Name())
	if err != nil {
		logErrorAndExit("unable to get batch report absolute path: ", err)
	}

	return abs
}

func init() {
	batchCmd.Flags().StringP("format", "f", "html", "Output format for files (html, json, markdown, or csv)")
	batchCmd.Flags().Float64P("threshold", "t", 0, "Cost threshold for alerting on expensive queries (0 = disabled)")
	batchCmd.Flags().BoolP("recommend-indexes", "i", false, "Recommend indexes based on query execution plans")
	batchCmd.Flags().Float64("index-threshold", 100.0, "Minimum operation cost to trigger index recommendations")
	batchCmd.Flags().BoolP("combined", "c", false, "Generate a single combined report instead of individual files")
	batchCmd.Flags().StringP("output-dir", "o", "", "Directory to save output files (default: current directory)")
	batchCmd.Flags().Bool("continue-on-error", true, "Continue processing remaining queries if one fails")
	rootCmd.AddCommand(batchCmd)
}