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
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [SQL_QUERY]",
	Short: "Create execution plan for a given SQL query",
	Long:  "Generate an execution plan and access it from a remote server or the file system",
	Args:  cobra.MaximumNArgs(1),
	Run:   runExplain,
}

func runExplain(cmd *cobra.Command, args []string) {
	// Get query from file flag, stdin, or argument
	query, err := getQueryInput(cmd, args)
	if err != nil {
		logErrorAndExit("Failed to get query input: ", err)
	}

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
		case "markdown":
			fmt.Println("💾 Generating Markdown report...")
			fileName = writeMarkdownPlan(plan, query, title, costInfo)
		case "csv":
			fmt.Println("💾 Saving as CSV...")
			fileName = writeCSVPlan(plan, query, title, costInfo)
		default:
			logErrorAndExit("Invalid format specified", fmt.Errorf("supported formats: html, json, markdown, csv"))
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

// getQueryInput retrieves the SQL query from various input sources
// Priority: --file flag > STDIN > --editor flag > interactive prompt > command argument
func getQueryInput(cmd *cobra.Command, args []string) (string, error) {
	// 1. Check if --file flag is provided
	filePath, _ := cmd.Flags().GetString("file")
	if filePath != "" {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
		}
		query := strings.TrimSpace(string(content))
		if query == "" {
			return "", fmt.Errorf("file %s is empty", filePath)
		}
		return query, nil
	}

	// 2. Try to read from STDIN (piped input)
	stat, err := os.Stdin.Stat()
	if err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
		bytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read from stdin: %w", err)
		}
		query := strings.TrimSpace(string(bytes))
		if query == "" {
			return "", fmt.Errorf("stdin input is empty")
		}
		return query, nil
	}

	// 3. Check if query is provided as command argument
	if len(args) > 0 {
		return args[0], nil
	}

	// 4. Check if --editor flag is set
	useEditor, _ := cmd.Flags().GetBool("editor")
	if useEditor {
		return getQueryFromEditor()
	}

	// 5. Default: Interactive multi-line prompt
	return getQueryFromPrompt()
}

// getQueryFromEditor opens the user's default editor to input the query
func getQueryFromEditor() (string, error) {
	// Get editor from environment, fallback to vim
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vim" // fallback
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "pgexplain_query_*.sql")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// Write placeholder text
	placeholder := "-- Enter your SQL query below, then save and close the editor\n\n"
	if err := os.WriteFile(tmpPath, []byte(placeholder), 0644); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	// Open editor
	fmt.Printf("\n✏️  Opening editor: %s\n", editor)
	fmt.Println("💡 Write your query, save, and close the editor to continue...")

	cmd := exec.Command(editor, tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor exited with error: %w", err)
	}

	// Read the query from the temp file
	content, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", fmt.Errorf("failed to read temp file: %w", err)
	}

	// Remove comments and trim
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

// getQueryFromPrompt prompts the user to paste/type their query interactively
func getQueryFromPrompt() (string, error) {
	fmt.Println("\n📝 Enter your SQL query (paste or type, press Ctrl+D when done):")
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

	password := os.Getenv("PGPASSWORD")
	if password == "" && config.Database.Password != "" {
		password = config.Database.Password
	}

	sql := fmt.Sprintf("EXPLAIN (ANALYSE, BUFFERS) %s", query)

	execution := exec.Command("psql", "-c", sql, "-U", user, "-d", database, "-h", host)

	// Set PGPASSWORD in the command's environment if available
	// This is more secure than passing it as a command argument
	if password != "" {
		execution.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", password))
	}

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
	analyzeCmd.Flags().StringP("format", "f", "html", "Output format for local files (html, json, markdown, or csv)")
	analyzeCmd.Flags().StringP("file", "F", "", "Read SQL query from file")
	analyzeCmd.Flags().BoolP("editor", "e", false, "Open $EDITOR to write/paste query")
	analyzeCmd.Flags().Float64P("threshold", "t", 0, "Cost threshold for alerting on expensive queries (0 = disabled)")
	analyzeCmd.Flags().BoolP("recommend-indexes", "i", false, "Recommend indexes based on query execution plan")
	analyzeCmd.Flags().Float64("index-threshold", 100.0, "Minimum operation cost to trigger index recommendations")
	rootCmd.AddCommand(analyzeCmd)
}
