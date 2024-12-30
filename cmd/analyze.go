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

	plan, err := generateExecutionPlan(query)
	if err != nil {
		logErrorAndExit("Error generating execution plan: ", err)
	}

	title := generateTitle()

	remoteFlag, _ := cmd.Flags().GetBool("remote")

	if remoteFlag {
		remoteURL := uploadPlan(plan, query, title)
		fmt.Printf("Access the plan from the remote URL: %s\n", remoteURL)
	} else {
		fileName := writePlan(plan, query, title)
		fmt.Printf("Access the plan from the file system: %s\n", fileName)
	}
}

func generateExecutionPlan(query string) (string, error) {
	// Define the psql command and its arguments. Ensure your psql configuration is properly initialized
	// before executing the command. For more details, @see the PostgreSQL environment variables : https://www.postgresql.org/docs/current/libpq-envars.html

	// It is not recommended to store passwords directly in the application. Instead, use a .pgpass configuration file.
	// For example, you can create a .pgpass file with the following content:
	// echo "$PGHOST:5432:$PGDATABASE:$PGUSER:$PGPASSWORD" > ~/.pgpass
	// Refer to the .pgpass file documentation for more information: @see https://www.postgresql.org/docs/current/libpq-pgpass.html
	user := os.Getenv("PGUSER")
	database := os.Getenv("PGDATABASE")
	host := os.Getenv("PGHOST")

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
	rootCmd.AddCommand(analyzeCmd)
}
