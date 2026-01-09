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
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Create execution plan for a given SQL query",
	Long:  "Generate an execution plan and access it from a remote server or the file system",
	Args:  cobra.ExactArgs(1),
	Run:   runExplain,
}

func runExplain(cmd *cobra.Command, args []string) {
	query := args[0]

	// Load configuration
	config, _ := loadConfig()

	// Get flag values, using config defaults if flags not explicitly set
	threshold, _ := cmd.Flags().GetFloat64("threshold")
	if !cmd.Flags().Changed("threshold") && config.Defaults.Threshold > 0 {
		threshold = config.Defaults.Threshold
	}

	remoteFlag, _ := cmd.Flags().GetBool("remote")
	if !cmd.Flags().Changed("remote") {
		remoteFlag = config.Defaults.Remote
	}

	format, _ := cmd.Flags().GetString("format")
	if !cmd.Flags().Changed("format") && config.Defaults.Format != "" {
		format = config.Defaults.Format
	}

	// Show friendly start message
	fmt.Println("\n🔍 Analyzing your query...")
	fmt.Printf("📊 Output format: %s\n", format)
	if threshold > 0 {
		fmt.Printf("⚡ Cost threshold: %.0f\n", threshold)
	}
	fmt.Println()

	plan, err := generateExecutionPlan(query, config)
	if err != nil {
		fmt.Println("❌ Failed to analyze query")
		logErrorAndExit("Error: ", err)
	}

	fmt.Println("✅ Query analysis complete!")
	fmt.Println()

	title := generateTitle()

	// Cost analysis
	var costInfo *CostInfo
	if threshold > 0 {
		costInfo = parseCost(plan, threshold)
		if costInfo.ExceedsLimit {
			displayCostAlert(costInfo)
		} else {
			fmt.Printf("✨ Great! Query cost (%.2f) is below threshold (%.0f)\n\n", costInfo.TotalCost, threshold)
		}
	}

	// Index recommendations
	recommendIndexes, _ := cmd.Flags().GetBool("recommend-indexes")
	if recommendIndexes {
		indexThreshold, _ := cmd.Flags().GetFloat64("index-threshold")
		indexInfo := analyzeIndexOpportunities(plan, indexThreshold)
		if indexInfo.TotalFound > 0 {
			displayIndexRecommendations(indexInfo)
		} else {
			fmt.Printf("✨ No index recommendations (all operations below threshold of %.0f)\n\n", indexThreshold)
		}
	}

	if remoteFlag {
		fmt.Println("☁️  Uploading to remote server...")
		remoteURL := uploadPlan(plan, query, title)
		fmt.Println("\n" + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("🌐 Remote URL (share with your team):")
		fmt.Printf("   %s\n", remoteURL)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	} else {
		var fileName string
		switch format {
		case "json":
			fmt.Println("💾 Saving as JSON...")
			fileName = writeJSONPlan(plan, query, title, costInfo)
		case "html":
			fmt.Println("💾 Generating interactive HTML report...")
			fileName = writePlan(plan, query, title)
		default:
			logErrorAndExit("Invalid format specified", fmt.Errorf("supported formats: html, json"))
		}

		fmt.Println("\n" + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("📁 Plan saved successfully!")
		fmt.Printf("   %s\n", fileName)
		if format == "html" {
			fmt.Println("\n💡 Tip: Open this file in your browser to view the interactive plan")
		}
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	}
}

func generateExecutionPlan(query string, config *Config) (string, error) {
	// Define the psql command and its arguments. Ensure your psql configuration is properly initialized
	// before executing the command. For more details, @see the PostgreSQL environment variables : https://www.postgresql.org/docs/current/libpq-envars.html

	// It is not recommended to store passwords directly in the application. Instead, use a .pgpass configuration file.
	// For example, you can create a .pgpass file with the following content:
	// echo "$PGHOST:5432:$PGDATABASE:$PGUSER:$PGPASSWORD" > ~/.pgpass
	// Refer to the .pgpass file documentation for more information: @see https://www.postgresql.org/docs/current/libpq-pgpass.html

	// Use environment variables first, fall back to config
	user := os.Getenv("PGUSER")
	if user == "" && config.Database.User != "" {
		user = config.Database.User
	}

	database := os.Getenv("PGDATABASE")
	if database == "" && config.Database.Database != "" {
		database = config.Database.Database
	}

	host := os.Getenv("PGHOST")
	if host == "" && config.Database.Host != "" {
		host = config.Database.Host
	}

	sql := fmt.Sprintf("EXPLAIN (ANALYSE, BUFFERS) %s", query)

	execution := exec.Command("psql", "-c", sql, "-U", user, "-d", database, "-h", host)

	plan, err := execution.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("unable to analyze the query: %w", err)
	}

	return string(plan), nil
}

func generateTitle() string {
	currentTime := time.Now()
	return fmt.Sprintf("Plan_Created_on_%s_%dth_%d_%02d:%02d:%02d",
		currentTime.Month(),
		currentTime.Day(),
		currentTime.Year(),
		currentTime.Hour(),
		currentTime.Minute(),
		currentTime.Second(),
	)
}

func logErrorAndExit(message string, err error) {
	log.Fatalf("%s: %v\n", message, err)
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pgexplain.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	analyzeCmd.Flags().BoolP("remote", "r", false, "Send the execution plan to a remote server to share with your individuals")
	analyzeCmd.Flags().StringP("format", "f", "html", "Output format for local files (html or json)")
	analyzeCmd.Flags().Float64P("threshold", "t", 0, "Cost threshold for alerting on expensive queries (0 = disabled)")
	analyzeCmd.Flags().BoolP("recommend-indexes", "i", false, "Recommend indexes based on query execution plan")
	analyzeCmd.Flags().Float64("index-threshold", 100.0, "Minimum operation cost to trigger index recommendations")
	rootCmd.AddCommand(analyzeCmd)
}
