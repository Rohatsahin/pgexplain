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
	"regexp"
	"strconv"
	"strings"
)

type CostInfo struct {
	TotalCost      float64
	ExpensiveOps   []ExpensiveOperation
	ExceedsLimit   bool
	ThresholdValue float64
}

type ExpensiveOperation struct {
	Operation string
	Cost      float64
	Line      string
}

// parseCost extracts cost information from a PostgreSQL EXPLAIN plan
func parseCost(plan string, threshold float64) *CostInfo {
	costInfo := &CostInfo{
		TotalCost:      0,
		ExpensiveOps:   []ExpensiveOperation{},
		ExceedsLimit:   false,
		ThresholdValue: threshold,
	}

	// Regex to match cost in format: cost=X..Y
	costRegex := regexp.MustCompile(`cost=(\d+\.?\d*)\.\.(\d+\.?\d*)`)

	lines := strings.Split(plan, "\n")
	for _, line := range lines {
		matches := costRegex.FindStringSubmatch(line)
		if len(matches) >= 3 {
			totalCost, err := strconv.ParseFloat(matches[2], 64)
			if err != nil {
				continue
			}

			// Track the highest cost as the total query cost
			if totalCost > costInfo.TotalCost {
				costInfo.TotalCost = totalCost
			}

			// Identify expensive operations
			if totalCost >= threshold {
				operation := extractOperationType(line)
				expensiveOp := ExpensiveOperation{
					Operation: operation,
					Cost:      totalCost,
					Line:      strings.TrimSpace(line),
				}
				costInfo.ExpensiveOps = append(costInfo.ExpensiveOps, expensiveOp)
			}
		}
	}

	if costInfo.TotalCost >= threshold {
		costInfo.ExceedsLimit = true
	}

	return costInfo
}

// extractOperationType extracts the operation type from an EXPLAIN line
func extractOperationType(line string) string {
	trimmed := strings.TrimSpace(line)

	// Common operation types in PostgreSQL
	operations := []string{
		"Seq Scan", "Index Scan", "Index Only Scan", "Bitmap Heap Scan",
		"Bitmap Index Scan", "Nested Loop", "Hash Join", "Merge Join",
		"Sort", "Aggregate", "Hash", "Materialize", "Gather", "Parallel Seq Scan",
	}

	for _, op := range operations {
		if strings.Contains(trimmed, op) {
			return op
		}
	}

	// If no specific operation found, extract the first few words
	parts := strings.Fields(trimmed)
	if len(parts) >= 2 {
		return strings.Join(parts[:2], " ")
	} else if len(parts) == 1 {
		return parts[0]
	}

	return "Unknown Operation"
}

// displayCostAlert prints cost threshold alerts to the user
func displayCostAlert(costInfo *CostInfo) {
	if !costInfo.ExceedsLimit {
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Printf("⚠️  COST THRESHOLD ALERT\n")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("Query Cost: %.2f (Threshold: %.2f)\n", costInfo.TotalCost, costInfo.ThresholdValue)
	fmt.Printf("Status: EXCEEDS THRESHOLD by %.2f\n", costInfo.TotalCost-costInfo.ThresholdValue)

	if len(costInfo.ExpensiveOps) > 0 {
		fmt.Printf("\nExpensive Operations Found: %d\n", len(costInfo.ExpensiveOps))
		fmt.Println(strings.Repeat("-", 70))
		for i, op := range costInfo.ExpensiveOps {
			fmt.Printf("%d. %s (Cost: %.2f)\n", i+1, op.Operation, op.Cost)
			fmt.Printf("   %s\n", op.Line)
		}
	}

	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("💡 Consider: Adding indexes, optimizing joins, or limiting result sets\n")
}