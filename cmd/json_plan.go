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
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type PlanOutput struct {
	Title         string     `json:"title"`
	Query         string     `json:"query"`
	ExecutionPlan string     `json:"execution_plan"`
	GeneratedAt   time.Time  `json:"generated_at"`
	CostAnalysis  *CostInfo  `json:"cost_analysis,omitempty"`
}

// writeJSONPlan generates a JSON file with the execution plan and query.
// It returns the absolute path of the generated file.
func writeJSONPlan(plan, query, title string, costInfo *CostInfo) string {
	name := title + ".json"
	data := PlanOutput{
		Title:         title,
		Query:         query,
		ExecutionPlan: plan,
		GeneratedAt:   time.Now(),
		CostAnalysis:  costInfo,
	}

	file, err := os.Create(name)
	if err != nil {
		logErrorAndExit("unable to create JSON plan file: ", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	err = encoder.Encode(data)
	if err != nil {
		logErrorAndExit("unable to encode plan to JSON: ", err)
	}

	abs, err := filepath.Abs(file.Name())
	if err != nil {
		logErrorAndExit("unable to get plan file absolute path: ", err)
	}

	return abs
}

// writeJSONToFile writes any data structure to a JSON file.
// It returns the absolute path of the generated file.
func writeJSONToFile(fileName string, data interface{}) string {
	file, err := os.Create(fileName)
	if err != nil {
		logErrorAndExit("unable to create JSON file: ", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	err = encoder.Encode(data)
	if err != nil {
		logErrorAndExit("unable to encode data to JSON: ", err)
	}

	abs, err := filepath.Abs(file.Name())
	if err != nil {
		logErrorAndExit("unable to get file absolute path: ", err)
	}

	return abs
}