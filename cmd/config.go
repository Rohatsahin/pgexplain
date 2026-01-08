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
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Defaults struct {
		Format    string  `yaml:"format"`
		Threshold float64 `yaml:"threshold"`
		Remote    bool    `yaml:"remote"`
	} `yaml:"defaults"`
	Database struct {
		Host     string `yaml:"host"`
		User     string `yaml:"user"`
		Database string `yaml:"database"`
	} `yaml:"database"`
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage pgexplain configuration",
	Long:  "Create and manage configuration files for pgexplain",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a default configuration file",
	Long:  "Generate a .pgexplainrc configuration file in your home directory",
	Run:   runConfigInit,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	Long:  "Show the current configuration settings",
	Run:   runConfigShow,
}

func runConfigInit(cmd *cobra.Command, args []string) {
	configPath := getConfigPath()

	fmt.Println("\n⚙️  Initializing pgexplain configuration...")
	fmt.Println()

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("⚠️  Configuration file already exists at:\n   %s\n\n", configPath)
		fmt.Print("Overwrite? (y/n): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("❌ Configuration creation cancelled.\n")
			return
		}
	}

	// Create default config
	defaultConfig := Config{}
	defaultConfig.Defaults.Format = "html"
	defaultConfig.Defaults.Threshold = 0
	defaultConfig.Defaults.Remote = false
	defaultConfig.Database.Host = os.Getenv("PGHOST")
	defaultConfig.Database.User = os.Getenv("PGUSER")
	defaultConfig.Database.Database = os.Getenv("PGDATABASE")

	// If env vars are empty, use placeholder values
	if defaultConfig.Database.Host == "" {
		defaultConfig.Database.Host = "localhost"
	}
	if defaultConfig.Database.User == "" {
		defaultConfig.Database.User = "postgres"
	}
	if defaultConfig.Database.Database == "" {
		defaultConfig.Database.Database = "mydb"
	}

	// Add comments to make it more user-friendly
	configContent := `# PG Explain Configuration File
# This file contains default settings for pgexplain
# You can override these settings using command-line flags

# Default settings for analyze command
defaults:
  format: html      # Output format: html or json
  threshold: 0      # Cost threshold for alerts (0 = disabled)
  remote: false     # Upload to remote server by default

# Database connection settings
# These override environment variables (PGHOST, PGUSER, PGDATABASE)
database:
  host: ` + defaultConfig.Database.Host + `
  user: ` + defaultConfig.Database.User + `
  database: ` + defaultConfig.Database.Database + `

# Note: Use .pgpass file for password management
# For more info: https://www.postgresql.org/docs/current/libpq-pgpass.html
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		fmt.Println("❌ Failed to create configuration file")
		logErrorAndExit("Error: ", err)
	}

	fmt.Println("✅ Configuration file created successfully!\n")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📝 Configuration file location:")
	fmt.Printf("   %s\n\n", configPath)
	fmt.Println("💡 Next steps:")
	fmt.Println("   1. Edit the file to customize your settings")
	fmt.Println("   2. Run 'pg_explain config show' to verify")
	fmt.Println("   3. Start using pgexplain with your defaults!")
	fmt.Println("\n   Note: Command-line flags will override config settings")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
}

func runConfigShow(cmd *cobra.Command, args []string) {
	config, configPath := loadConfig()

	if configPath == "" {
		fmt.Println("\n❌ No configuration file found.\n")
		fmt.Println("💡 Create one by running:")
		fmt.Println("   pg_explain config init\n")
		return
	}

	fmt.Println("\n⚙️  Current Configuration")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📁 File: %s\n", configPath)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	fmt.Println("📊 Defaults:")
	fmt.Printf("   Format:      %s\n", config.Defaults.Format)
	fmt.Printf("   Threshold:   %.0f\n", config.Defaults.Threshold)
	fmt.Printf("   Remote:      %v\n", config.Defaults.Remote)

	fmt.Println("\n🗄️  Database:")
	fmt.Printf("   Host:        %s\n", config.Database.Host)
	fmt.Printf("   User:        %s\n", config.Database.User)
	fmt.Printf("   Database:    %s\n", config.Database.Database)

	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("💡 Note: Command-line flags will override these settings")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
}

func getConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".pgexplainrc"
	}
	return filepath.Join(home, ".pgexplainrc")
}

func loadConfig() (*Config, string) {
	// Try to load from home directory
	configPath := getConfigPath()
	config := &Config{}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Try current directory as fallback
		localConfigPath := ".pgexplainrc"
		if _, err := os.Stat(localConfigPath); os.IsNotExist(err) {
			// No config file found, return defaults
			config.Defaults.Format = "html"
			config.Defaults.Threshold = 0
			config.Defaults.Remote = false
			return config, ""
		}
		configPath = localConfigPath
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		// Return defaults if can't read
		config.Defaults.Format = "html"
		config.Defaults.Threshold = 0
		config.Defaults.Remote = false
		return config, ""
	}

	// Parse YAML
	err = yaml.Unmarshal(data, config)
	if err != nil {
		fmt.Printf("Warning: Failed to parse config file: %v\n", err)
		// Return defaults on parse error
		config.Defaults.Format = "html"
		config.Defaults.Threshold = 0
		config.Defaults.Remote = false
		return config, ""
	}

	return config, configPath
}

func init() {
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	rootCmd.AddCommand(configCmd)
}