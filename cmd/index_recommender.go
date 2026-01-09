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

// IndexRecommendation represents a single index recommendation
type IndexRecommendation struct {
	TableName       string
	Columns         []string
	IndexType       string
	Reason          string
	OperationType   string
	OperationCost   float64
	CreateStatement string
	Priority        int
}

// IndexRecommendationInfo aggregates all recommendations
type IndexRecommendationInfo struct {
	Recommendations []IndexRecommendation
	TotalFound      int
	HighPriority    int
	ThresholdUsed   float64
}

// OperationContext holds parsed information about a single EXPLAIN line
type OperationContext struct {
	Line          string
	OperationType string
	TableName     string
	FilterColumns []string
	JoinColumns   []string
	SortColumns   []string
	Cost          float64
	RowsEstimate  int64
}

// Regex patterns for parsing EXPLAIN output
var (
	tableNameRegex    = regexp.MustCompile(`(?:Seq Scan|Parallel Seq Scan|Index Scan|Index Only Scan|Bitmap Heap Scan)\s+on\s+(\w+)`)
	filterRegex       = regexp.MustCompile(`Filter:\s*\(([^)]+(?:\([^)]*\)[^)]*)*)\)`)
	filterColumnRegex = regexp.MustCompile(`\b(\w+)\s*(?:=|>|<|>=|<=|!=|<>|~~|LIKE|IN|IS)`)
	hashCondRegex     = regexp.MustCompile(`Hash Cond:\s*\(([^)]+)\)`)
	mergeCondRegex    = regexp.MustCompile(`Merge Cond:\s*\(([^)]+)\)`)
	sortKeyRegex      = regexp.MustCompile(`Sort Key:\s*(.+)`)
	costRegex         = regexp.MustCompile(`cost=(\d+\.?\d*)\.\.(\d+\.?\d*)`)
	rowsRegex         = regexp.MustCompile(`rows=(\d+)`)
	joinColumnRegex   = regexp.MustCompile(`(\w+)\.(\w+)\s*=\s*(\w+)\.(\w+)`)
)

// analyzeIndexOpportunities is the main entry point for index recommendation analysis
func analyzeIndexOpportunities(plan string, threshold float64) *IndexRecommendationInfo {
	contexts := parseExplainForIndexes(plan, threshold)
	return generateIndexRecommendations(contexts, threshold)
}

// parseExplainForIndexes parses the EXPLAIN output and extracts operation contexts
func parseExplainForIndexes(plan string, threshold float64) []OperationContext {
	contexts := []OperationContext{}
	lines := strings.Split(plan, "\n")

	for i, line := range lines {
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Parse cost
		costMatches := costRegex.FindStringSubmatch(line)
		if len(costMatches) < 3 {
			continue
		}
		cost, err := strconv.ParseFloat(costMatches[2], 64)
		if err != nil {
			continue
		}

		// Only analyze operations above threshold
		if cost < threshold {
			continue
		}

		context := OperationContext{
			Line: strings.TrimSpace(line),
			Cost: cost,
		}

		// Extract operation type
		context.OperationType = extractOperationType(line)

		// Extract table name
		if tableMatches := tableNameRegex.FindStringSubmatch(line); len(tableMatches) > 1 {
			context.TableName = tableMatches[1]
		}

		// Extract row estimate
		if rowMatches := rowsRegex.FindStringSubmatch(line); len(rowMatches) > 1 {
			context.RowsEstimate, _ = strconv.ParseInt(rowMatches[1], 10, 64)
		}

		// Look ahead for Filter, Hash Cond, Sort Key in next few lines (indented child lines)
		for j := i + 1; j < len(lines) && j < i+5; j++ {
			nextLine := lines[j]

			// Check if still indented (child of current operation)
			if !strings.HasPrefix(nextLine, "  ") && !strings.HasPrefix(nextLine, "\t") {
				break
			}

			// Extract filter columns
			if filterMatches := filterRegex.FindStringSubmatch(nextLine); len(filterMatches) > 1 {
				filterExpr := filterMatches[1]
				columnMatches := filterColumnRegex.FindAllStringSubmatch(filterExpr, -1)
				for _, match := range columnMatches {
					if len(match) > 1 {
						// Avoid duplicates
						col := match[1]
						found := false
						for _, existing := range context.FilterColumns {
							if existing == col {
								found = true
								break
							}
						}
						if !found {
							context.FilterColumns = append(context.FilterColumns, col)
						}
					}
				}
			}

			// Extract hash join columns
			if hashMatches := hashCondRegex.FindStringSubmatch(nextLine); len(hashMatches) > 1 {
				joinExpr := hashMatches[1]
				if joinColMatches := joinColumnRegex.FindStringSubmatch(joinExpr); len(joinColMatches) > 4 {
					// Store as "table.column" pairs
					context.JoinColumns = append(context.JoinColumns,
						joinColMatches[1]+"."+joinColMatches[2],
						joinColMatches[3]+"."+joinColMatches[4])
				}
			}

			// Extract merge join columns
			if mergeMatches := mergeCondRegex.FindStringSubmatch(nextLine); len(mergeMatches) > 1 {
				joinExpr := mergeMatches[1]
				if joinColMatches := joinColumnRegex.FindStringSubmatch(joinExpr); len(joinColMatches) > 4 {
					context.JoinColumns = append(context.JoinColumns,
						joinColMatches[1]+"."+joinColMatches[2],
						joinColMatches[3]+"."+joinColMatches[4])
				}
			}

			// Extract sort keys
			if sortMatches := sortKeyRegex.FindStringSubmatch(nextLine); len(sortMatches) > 1 {
				sortKeys := strings.Split(sortMatches[1], ",")
				for _, key := range sortKeys {
					key = strings.TrimSpace(key)
					// Remove DESC/ASC keywords
					key = strings.TrimSuffix(key, " DESC")
					key = strings.TrimSuffix(key, " ASC")
					// Handle table.column or just column
					if strings.Contains(key, ".") {
						parts := strings.Split(key, ".")
						if len(parts) >= 2 {
							context.SortColumns = append(context.SortColumns, parts[1])
						}
					} else {
						context.SortColumns = append(context.SortColumns, key)
					}
				}
			}
		}

		contexts = append(contexts, context)
	}

	return contexts
}

// generateIndexRecommendations analyzes operation contexts and generates recommendations
func generateIndexRecommendations(contexts []OperationContext, threshold float64) *IndexRecommendationInfo {
	info := &IndexRecommendationInfo{
		Recommendations: []IndexRecommendation{},
		ThresholdUsed:   threshold,
	}

	// Track recommendations to avoid duplicates (key: "table:column1,column2")
	seen := make(map[string]bool)

	for _, ctx := range contexts {
		// Skip if no table name identified
		if ctx.TableName == "" {
			continue
		}

		// Rule 1: Sequential Scan with Filter -> Recommend index on filtered columns
		if strings.Contains(ctx.OperationType, "Seq Scan") && len(ctx.FilterColumns) > 0 {
			for _, col := range ctx.FilterColumns {
				rec := IndexRecommendation{
					TableName:     ctx.TableName,
					Columns:       []string{col},
					IndexType:     "BTREE",
					Reason:        fmt.Sprintf("Sequential scan with filter on '%s'", col),
					OperationType: ctx.OperationType,
					OperationCost: ctx.Cost,
					Priority:      calculatePriority(ctx.Cost, ctx.RowsEstimate, "filter"),
				}
				rec.CreateStatement = formatCreateIndexStatement(rec)

				key := fmt.Sprintf("%s:%s", rec.TableName, strings.Join(rec.Columns, ","))
				if validateRecommendation(rec) && !seen[key] {
					info.Recommendations = append(info.Recommendations, rec)
					seen[key] = true
				}
			}
		}

		// Rule 2: Hash/Merge Join conditions -> Recommend indexes on join columns
		if (strings.Contains(ctx.OperationType, "Hash Join") ||
			strings.Contains(ctx.OperationType, "Merge Join")) &&
			len(ctx.JoinColumns) > 0 {

			for _, joinCol := range ctx.JoinColumns {
				// Parse "table.column"
				parts := strings.Split(joinCol, ".")
				if len(parts) != 2 {
					continue
				}
				tableName := parts[0]
				columnName := parts[1]

				rec := IndexRecommendation{
					TableName:     tableName,
					Columns:       []string{columnName},
					IndexType:     "BTREE",
					Reason:        fmt.Sprintf("Join condition on '%s.%s'", tableName, columnName),
					OperationType: ctx.OperationType,
					OperationCost: ctx.Cost,
					Priority:      calculatePriority(ctx.Cost, ctx.RowsEstimate, "join"),
				}
				rec.CreateStatement = formatCreateIndexStatement(rec)

				key := fmt.Sprintf("%s:%s", rec.TableName, strings.Join(rec.Columns, ","))
				if validateRecommendation(rec) && !seen[key] {
					info.Recommendations = append(info.Recommendations, rec)
					seen[key] = true
				}
			}
		}

		// Rule 3: Expensive Sort -> Recommend index on sort columns
		if strings.Contains(ctx.OperationType, "Sort") && len(ctx.SortColumns) > 0 && ctx.TableName != "" {
			// Multi-column index for compound sort keys
			rec := IndexRecommendation{
				TableName:     ctx.TableName,
				Columns:       ctx.SortColumns,
				IndexType:     "BTREE",
				Reason:        fmt.Sprintf("Expensive sort operation on %s", strings.Join(ctx.SortColumns, ", ")),
				OperationType: ctx.OperationType,
				OperationCost: ctx.Cost,
				Priority:      calculatePriority(ctx.Cost, ctx.RowsEstimate, "sort"),
			}
			rec.CreateStatement = formatCreateIndexStatement(rec)

			key := fmt.Sprintf("%s:%s", rec.TableName, strings.Join(rec.Columns, ","))
			if validateRecommendation(rec) && !seen[key] {
				info.Recommendations = append(info.Recommendations, rec)
				seen[key] = true
			}
		}
	}

	// Sort by priority (descending) then by cost (descending)
	sortRecommendations(info.Recommendations)

	info.TotalFound = len(info.Recommendations)
	for _, rec := range info.Recommendations {
		if rec.Priority >= 4 {
			info.HighPriority++
		}
	}

	return info
}

// calculatePriority assigns priority based on cost, rows, and operation type
func calculatePriority(cost float64, rows int64, operationType string) int {
	// Base priority on cost
	priority := 1

	if cost > 10000 {
		priority = 5
	} else if cost > 5000 {
		priority = 4
	} else if cost > 1000 {
		priority = 3
	} else if cost > 500 {
		priority = 2
	}

	// Boost priority for high row counts (seq scans on large tables)
	if rows > 100000 {
		priority = minInt(5, priority+1)
	}

	// Joins are slightly higher priority than filters
	if operationType == "join" {
		priority = minInt(5, priority+1)
	}

	return priority
}

// formatCreateIndexStatement generates the CREATE INDEX SQL
func formatCreateIndexStatement(rec IndexRecommendation) string {
	// Generate a meaningful index name: idx_<table>_<col1>_<col2>
	indexName := fmt.Sprintf("idx_%s_%s", rec.TableName, strings.Join(rec.Columns, "_"))

	// Format columns
	columnList := strings.Join(rec.Columns, ", ")

	return fmt.Sprintf("CREATE INDEX %s ON %s USING %s (%s);",
		indexName,
		rec.TableName,
		rec.IndexType,
		columnList)
}

// validateRecommendation checks if a recommendation is valid
func validateRecommendation(rec IndexRecommendation) bool {
	// Must have table name
	if rec.TableName == "" {
		return false
	}

	// Must have at least one column
	if len(rec.Columns) == 0 {
		return false
	}

	// Exclude system tables/columns
	systemTables := []string{"pg_catalog", "information_schema"}
	for _, sys := range systemTables {
		if strings.HasPrefix(rec.TableName, sys) {
			return false
		}
	}

	// Valid column names (alphanumeric + underscore)
	validColumnName := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	for _, col := range rec.Columns {
		if !validColumnName.MatchString(col) {
			return false
		}
	}

	return true
}

// sortRecommendations sorts by priority (desc) then cost (desc)
func sortRecommendations(recommendations []IndexRecommendation) {
	// Bubble sort - simple implementation
	for i := 0; i < len(recommendations); i++ {
		for j := i + 1; j < len(recommendations); j++ {
			if recommendations[j].Priority > recommendations[i].Priority ||
				(recommendations[j].Priority == recommendations[i].Priority &&
					recommendations[j].OperationCost > recommendations[i].OperationCost) {
				recommendations[i], recommendations[j] = recommendations[j], recommendations[i]
			}
		}
	}
}

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// displayIndexRecommendations prints recommendations to console
func displayIndexRecommendations(info *IndexRecommendationInfo) {
	if info.TotalFound == 0 {
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Printf("🎯 INDEX RECOMMENDATIONS\n")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("Found: %d recommendations", info.TotalFound)
	if info.HighPriority > 0 {
		fmt.Printf(" (%d high priority)", info.HighPriority)
	}
	fmt.Println()
	fmt.Printf("Threshold: Operations with cost >= %.0f\n", info.ThresholdUsed)

	// Group by priority for better readability
	priorityGroups := make(map[int][]IndexRecommendation)
	for _, rec := range info.Recommendations {
		priorityGroups[rec.Priority] = append(priorityGroups[rec.Priority], rec)
	}

	// Display high priority first (5 down to 1)
	for priority := 5; priority >= 1; priority-- {
		recs := priorityGroups[priority]
		if len(recs) == 0 {
			continue
		}

		fmt.Println(strings.Repeat("-", 70))
		fmt.Printf("\n%s Priority %d %s\n", getPriorityEmoji(priority), priority, getPriorityLabel(priority))

		for i, rec := range recs {
			fmt.Printf("\n%d. Table: %s\n", i+1, rec.TableName)
			fmt.Printf("   Columns: %s\n", strings.Join(rec.Columns, ", "))
			fmt.Printf("   Reason: %s\n", rec.Reason)
			fmt.Printf("   Operation: %s (Cost: %.2f)\n", rec.OperationType, rec.OperationCost)
			fmt.Printf("   \n")
			fmt.Printf("   %s\n", rec.CreateStatement)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("💡 Tips:")
	fmt.Println("   • Test indexes on a development database first")
	fmt.Println("   • Monitor index usage with pg_stat_user_indexes")
	fmt.Println("   • Consider impact on INSERT/UPDATE performance")
	fmt.Println("   • Combine multiple single-column indexes into composite indexes where appropriate")
	fmt.Println(strings.Repeat("=", 70) + "\n")
}

// getPriorityEmoji returns emoji for priority level
func getPriorityEmoji(priority int) string {
	switch priority {
	case 5:
		return "🔴"
	case 4:
		return "🟠"
	case 3:
		return "🟡"
	case 2:
		return "🔵"
	default:
		return "⚪"
	}
}

// getPriorityLabel returns description for priority level
func getPriorityLabel(priority int) string {
	switch priority {
	case 5:
		return "(Critical - Very High Cost)"
	case 4:
		return "(High - Significant Impact)"
	case 3:
		return "(Medium - Moderate Impact)"
	case 2:
		return "(Low - Minor Impact)"
	default:
		return "(Minimal Impact)"
	}
}
