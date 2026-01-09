<p align="center">
  <img src="./assets/project_image.webp" alt="Project Overview" width="400" height="400" />
  <h3 align="center">PG Explain</h3>
  <p align="center">A command-line tool to analyze and visualize PostgreSQL database queries with pev2</p>
  <p align="center">
  <a href="https://opensource.org/licenses/Apache-2.0"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="Apache 2.0"></a>
</p>

---

## About The Project

PG Explain is a powerful command-line tool for analyzing and visualizing PostgreSQL query execution plans. Built with Go and Cobra, it provides an intuitive interface for generating execution plans with multiple output formats and intelligent cost analysis to help you optimize your database queries.

## Features

- **Configuration File Support**: Save your preferences in `.pgexplainrc` - perfect for non-developers
- **Query Comparison**: Compare two queries side-by-side to identify the most efficient approach
- **Multiple Output Formats**: Generate execution plans as interactive HTML or structured JSON
- **Cost Threshold Alerts**: Automatically detect and warn about expensive queries
- **Index Recommendations**: Get intelligent index suggestions based on query execution patterns
- **Remote Sharing**: Upload plans to Dalibo's pev2 service for easy sharing
- **Interactive Visualizations**: Beautiful HTML reports powered by [pev2](https://github.com/dalibo/pev2)
- **Cost Analysis**: Identify expensive operations and get optimization recommendations
- **Command-Oriented**: Built with Cobra for a structured and user-friendly CLI experience

---

## Quick Start

```bash
# Install
go install github.com/Rohatsahin/pgexplain@latest

# Set up PostgreSQL connection
export PGHOST=localhost
export PGUSER=myuser
export PGDATABASE=mydb

# Analyze a query (generates HTML file)
pg_explain analyze "SELECT * FROM users WHERE age > 25"

# Compare two queries to find the most efficient one
pg_explain compare "SELECT * FROM orders WHERE user_id = 123" "SELECT * FROM orders WHERE user_id IN (123)"

# Analyze with cost threshold warning
pg_explain analyze -t 1000 "SELECT * FROM large_table"

# Get index recommendations for query optimization
pg_explain analyze -i "SELECT * FROM users WHERE age > 25"

# Generate JSON output with cost analysis
pg_explain analyze -f json -t 500 "SELECT * FROM orders JOIN users ON orders.user_id = users.id"
```

---

## Installation

### Option 1: Install via Go (Recommended)

```bash
go install github.com/Rohatsahin/pgexplain@latest
```

This installs the binary to `$GOPATH/bin` (usually `~/go/bin`).

### Option 2: Build from Source

```bash
git clone https://github.com/Rohatsahin/pgexplain.git
cd pgexplain
go build -o pg_explain
```

#### Add to PATH (Linux/macOS)

```bash
sudo mv pg_explain /usr/local/bin/
```

#### Add to PATH (Windows)

1. Copy `pg_explain.exe` to `C:\Program Files\pg_explain\`
2. Add the directory to your PATH:
   - Right-click **This PC** → **Properties** → **Advanced system settings**
   - **Environment Variables** → **Path** → **Edit** → **New**
   - Add `C:\Program Files\pg_explain\`

### Verify Installation

```bash
pg_explain --help
```

---

## Configuration

### Configuration File (.pgexplainrc)

**Recommended for non-developers**: Create a configuration file to avoid typing flags every time.

#### Create Configuration File

```bash
pg_explain config init
```

**Output:**
```
⚙️  Initializing pgexplain configuration...

✅ Configuration file created successfully!

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📝 Configuration file location:
   ~/.pgexplainrc

💡 Next steps:
   1. Edit the file to customize your settings
   2. Run 'pg_explain config show' to verify
   3. Start using pgexplain with your defaults!

   Note: Command-line flags will override config settings
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

This creates `~/.pgexplainrc` with default settings:

```yaml
# PG Explain Configuration File
defaults:
  format: html      # Output format: html or json
  threshold: 0      # Cost threshold for alerts (0 = disabled)
  remote: false     # Upload to remote server by default

database:
  host: localhost
  user: postgres
  database: mydb
```

#### View Current Configuration

```bash
pg_explain config show
```

**Output:**
```
⚙️  Current Configuration
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📁 File: ~/.pgexplainrc
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📊 Defaults:
   Format:      html
   Threshold:   500
   Remote:      false

🗄️  Database:
   Host:        localhost
   User:        postgres
   Database:    mydb

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
💡 Note: Command-line flags will override these settings
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

#### How It Works

1. **Config file is optional** - Everything works without it
2. **Config provides defaults** - No need to type `-f json` every time
3. **Flags override config** - Command-line flags always take priority
4. **Location**: `~/.pgexplainrc` in your home directory

#### Example Workflow

```bash
# Set up once
pg_explain config init
# Edit ~/.pgexplainrc and set: format: json, threshold: 500

# Now these commands use your defaults
pg_explain analyze "SELECT * FROM users"  # Uses json format automatically
pg_explain analyze -f html "SELECT * FROM orders"  # Override with html
```

### PostgreSQL Connection

**Option 1: Configuration File (Recommended)**

Edit `~/.pgexplainrc`:
```yaml
database:
  host: localhost
  user: myuser
  database: mydb
```

**Option 2: Environment Variables**

```bash
export PGHOST=localhost
export PGUSER=myuser
export PGDATABASE=mydb
```

Or add them to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.).

**Priority**: Environment variables override config file settings.

### Secure Password Management

Use a `.pgpass` file instead of storing passwords in environment variables:

```bash
echo "localhost:5432:mydatabase:myuser:mypassword" > ~/.pgpass
chmod 600 ~/.pgpass
```

For more information, see the [PostgreSQL `.pgpass` documentation](https://www.postgresql.org/docs/current/libpq-pgpass.html).

---

## Usage

### Available Commands

Run the following command to see a list of available commands:

```bash
pg_explain --help
```

### Command Reference

#### `analyze` - Analyze SQL queries

```bash
pg_explain analyze [flags] "SQL_QUERY"
```

**Available Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--format` | `-f` | string | `html` | Output format: `html` or `json` |
| `--remote` | `-r` | bool | `false` | Upload plan to remote server for sharing |
| `--threshold` | `-t` | float | `0` | Cost threshold for alerting (0 = disabled) |
| `--recommend-indexes` | `-i` | bool | `false` | Recommend indexes based on query execution plan |
| `--index-threshold` | | float | `100` | Minimum operation cost to trigger index recommendations |

---

#### `compare` - Compare two SQL queries

```bash
pg_explain compare [flags] "QUERY1" "QUERY2"
```

**Available Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--format` | `-f` | string | `text` | Output format: `text` or `json` |

---

### Examples

#### 1. Basic Query Analysis (HTML Output)

Generate an interactive HTML visualization of your query plan:

```bash
pg_explain analyze "SELECT * FROM users WHERE age > 25"
```

**Output:**
```
🔍 Analyzing your query...
📊 Output format: html

✅ Query analysis complete!

💾 Generating interactive HTML report...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📁 Plan saved successfully!
   /path/to/Plan_Created_on_January_8th_2026_14:30:00.html

💡 Tip: Open this file in your browser to view the interactive plan
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

#### 2. JSON Output Format

Generate machine-readable JSON output for programmatic processing:

```bash
pg_explain analyze -f json "SELECT * FROM orders"
```

**Terminal Output:**
```
🔍 Analyzing your query...
📊 Output format: json

✅ Query analysis complete!

💾 Saving as JSON...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📁 Plan saved successfully!
   /path/to/Plan_Created_on_January_8th_2026_14:30:00.json
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**File Contents:**
```json
{
  "title": "Plan_Created_on_January_8th_2026_14:30:00",
  "query": "SELECT * FROM orders",
  "execution_plan": "Seq Scan on orders (cost=0.00..1250.50...)",
  "generated_at": "2026-01-08T14:30:00Z"
}
```

---

#### 3. Cost Threshold Alerts

Get warnings when queries exceed a cost threshold:

```bash
pg_explain analyze -t 1000 "SELECT * FROM large_table"
```

**Output with Alert:**
```
======================================================================
⚠️  COST THRESHOLD ALERT
======================================================================
Query Cost: 1250.50 (Threshold: 1000.00)
Status: EXCEEDS THRESHOLD by 250.50

Expensive Operations Found: 2
----------------------------------------------------------------------
1. Seq Scan (Cost: 1250.50)
   Seq Scan on large_table  (cost=0.00..1250.50 rows=50000 width=244)
2. Sort (Cost: 850.25)
   Sort  (cost=500.00..850.25 rows=10000 width=50)
======================================================================
💡 Consider: Adding indexes, optimizing joins, or limiting result sets

Access the plan from the file system: /path/to/Plan_Created_on_...html
```

---

#### 4. Combined: JSON + Cost Analysis

Generate JSON output with cost analysis data:

```bash
pg_explain analyze -f json -t 500 "SELECT o.*, u.name FROM orders o JOIN users u ON o.user_id = u.id"
```

**Terminal Output:**
```
🔍 Analyzing your query...
📊 Output format: json
⚡ Cost threshold: 500

✅ Query analysis complete!

======================================================================
⚠️  COST THRESHOLD ALERT
======================================================================
Query Cost: 850.25 (Threshold: 500.00)
Status: EXCEEDS THRESHOLD by 350.25

Expensive Operations Found: 1
----------------------------------------------------------------------
1. Hash Join (Cost: 850.25)
   Hash Join  (cost=125.00..850.25 rows=10000 width=100)
======================================================================
💡 Consider: Adding indexes, optimizing joins, or limiting result sets

💾 Saving as JSON...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📁 Plan saved successfully!
   /path/to/Plan_Created_on_January_8th_2026_14:30:00.json
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**File Contents:**
```json
{
  "title": "Plan_Created_on_January_8th_2026_14:30:00",
  "query": "SELECT o.*, u.name FROM orders o JOIN users u ON o.user_id = u.id",
  "execution_plan": "Hash Join (cost=125.00..850.25...)",
  "generated_at": "2026-01-08T14:30:00Z",
  "cost_analysis": {
    "TotalCost": 850.25,
    "ExpensiveOps": [
      {
        "Operation": "Hash Join",
        "Cost": 850.25,
        "Line": "Hash Join  (cost=125.00..850.25 rows=10000 width=100)"
      }
    ],
    "ExceedsLimit": true,
    "ThresholdValue": 500
  }
}
```

---

#### 5. Remote Sharing

Upload your execution plan to Dalibo's pev2 service for easy sharing:

```bash
pg_explain analyze --remote "SELECT * FROM products WHERE category = 'electronics'"
```

**Output:**
```
🔍 Analyzing your query...
📊 Output format: html

✅ Query analysis complete!

☁️  Uploading to remote server...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🌐 Remote URL (share with your team):
   https://explain.dalibo.com/plan/abc123def456
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

#### 6. Index Recommendations

Get intelligent index suggestions to optimize query performance:

```bash
pg_explain analyze --recommend-indexes "SELECT * FROM users WHERE age > 25 AND status = 'active'"
```

**Output:**
```
🔍 Analyzing your query...
📊 Output format: html

✅ Query analysis complete!

======================================================================
🎯 INDEX RECOMMENDATIONS
======================================================================
Found: 2 recommendations (1 high priority)
Threshold: Operations with cost >= 100.0
----------------------------------------------------------------------

🟠 Priority 4 (High - Significant Impact)

1. Table: users
   Columns: age
   Reason: Sequential scan with filter on 'age'
   Operation: Seq Scan (Cost: 5230.75)

   CREATE INDEX idx_users_age ON users USING BTREE (age);

----------------------------------------------------------------------

🟡 Priority 3 (Medium - Moderate Impact)

1. Table: users
   Columns: status
   Reason: Sequential scan with filter on 'status'
   Operation: Seq Scan (Cost: 1250.30)

   CREATE INDEX idx_users_status ON users USING BTREE (status);

======================================================================
💡 Tips:
   • Test indexes on a development database first
   • Monitor index usage with pg_stat_user_indexes
   • Consider impact on INSERT/UPDATE performance
   • Combine multiple single-column indexes into composite indexes where appropriate
======================================================================

💾 Generating interactive HTML report...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📁 Plan saved successfully!
   /path/to/Plan_Created_on_January_9th_2026_01:30:00.html
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**With Custom Threshold:**
```bash
pg_explain analyze -i --index-threshold 500 "SELECT * FROM orders JOIN users ON orders.user_id = users.id"
```

This will only recommend indexes for operations with cost >= 500.

**Combined with Cost Analysis:**
```bash
pg_explain analyze -t 1000 -i "SELECT * FROM large_table WHERE created_at > '2024-01-01'"
```

---

#### 7. Query Comparison

Compare two different query approaches to find the most efficient one:

```bash
pg_explain compare "SELECT * FROM orders WHERE status = 'pending'" "SELECT * FROM orders WHERE status IN ('pending')"
```

**Output:**
```
🔬 Starting query comparison...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔍 Analyzing Query 1...
✅ Query 1 complete!

🔍 Analyzing Query 2...
✅ Query 2 complete!

================================================================================
QUERY COMPARISON REPORT
================================================================================

Query 1:
  SELECT * FROM orders WHERE status = 'pending'
  Total Cost: 425.50
  Most Expensive Operation: Seq Scan (425.50)

--------------------------------------------------------------------------------

Query 2:
  SELECT * FROM orders WHERE status IN ('pending')
  Total Cost: 425.50
  Most Expensive Operation: Seq Scan (425.50)

================================================================================

COMPARISON RESULTS
--------------------------------------------------------------------------------
Winner: 🤝 Tie
Cost Difference: 0.00 (0.00%)

💡 Recommendation: Both queries have similar costs. Choose based on readability and maintainability.
================================================================================

DETAILED EXECUTION PLANS
--------------------------------------------------------------------------------

[Query 1 Execution Plan]
Seq Scan on orders  (cost=0.00..425.50 rows=1000 width=100)
  Filter: (status = 'pending'::text)
  ...

--------------------------------------------------------------------------------

[Query 2 Execution Plan]
Seq Scan on orders  (cost=0.00..425.50 rows=1000 width=100)
  Filter: (status = ANY ('{pending}'::text[]))
  ...

================================================================================
```

---

#### 8. Query Comparison with JSON Output

Get detailed comparison data in JSON format:

```bash
pg_explain compare -f json "SELECT * FROM users WHERE active = true" "SELECT * FROM users WHERE active IS true"
```

**Terminal Output:**
```
🔬 Starting query comparison...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔍 Analyzing Query 1...
✅ Query 1 complete!

🔍 Analyzing Query 2...
✅ Query 2 complete!

💾 Saving comparison as JSON...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📁 Comparison saved successfully!
   Comparison_Plan_Created_on_January_8th_2026_14:30:00.json

🏆 Winner: Query 2 (Cost diff: 24.61%)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**File Contents:**
```json
{
  "query1": "SELECT * FROM users WHERE active = true",
  "query2": "SELECT * FROM users WHERE active IS true",
  "plan1": "Seq Scan on users...",
  "plan2": "Seq Scan on users...",
  "cost_analysis1": {
    "TotalCost": 235.75,
    "ExpensiveOps": [...]
  },
  "cost_analysis2": {
    "TotalCost": 189.20,
    "ExpensiveOps": [...]
  },
  "winner": "Query 2",
  "cost_difference": 46.55,
  "cost_difference_percentage": 24.61,
  "recommendation": "Query 2 is more efficient. Consider using this approach."
}
```

---

### Real-World Use Cases

#### Detecting Slow Queries During Development

```bash
# Set a reasonable threshold for your application
pg_explain analyze -t 100 "SELECT * FROM users WHERE email LIKE '%@example.com%'"
```

Use this during development to catch performance issues early.

---

#### Comparing Query Plans

Generate JSON files for different query variations:

```bash
pg_explain analyze -f json "SELECT * FROM orders WHERE status = 'pending'" > query1.json
pg_explain analyze -f json "SELECT * FROM orders WHERE status IN ('pending')" > query2.json
```

Compare the cost analysis programmatically or use a diff tool.

---

#### Automated Performance Monitoring

Integrate into CI/CD pipelines:

```bash
#!/bin/bash
THRESHOLD=1000
OUTPUT=$(pg_explain analyze -f json -t $THRESHOLD "$QUERY")

if echo "$OUTPUT" | jq -e '.cost_analysis.ExceedsLimit == true' > /dev/null; then
  echo "Query exceeds cost threshold!"
  exit 1
fi
```

---

### Tips for Using Cost Thresholds

- **Development**: Use low thresholds (100-500) to catch issues early
- **Staging**: Use medium thresholds (500-2000) to validate optimizations
- **Production Monitoring**: Use higher thresholds (2000+) for critical alerts
- **Benchmark First**: Run your typical queries without thresholds to establish baselines

---

### Tips for Using Index Recommendations

- **Test First**: Always test recommended indexes on a development database before applying to production
- **Monitor Usage**: Use `pg_stat_user_indexes` to verify indexes are being used after creation
- **Consider Trade-offs**: Indexes improve read performance but can slow down INSERT/UPDATE operations
- **Composite Indexes**: Consider combining multiple single-column index recommendations into composite indexes
- **Adjust Threshold**: Use `--index-threshold` to focus on high-impact optimizations (e.g., `--index-threshold 500`)
- **Combine with Cost Analysis**: Run with both `-t` and `-i` flags to get comprehensive optimization insights
- **Priority Levels**: Focus on Priority 4-5 (High/Critical) recommendations first for maximum impact
- **Review Existing Indexes**: Check `pg_indexes` view to avoid creating duplicate indexes

---

## Troubleshooting

### Common Issues

**"psql: command not found"**
- Ensure PostgreSQL client tools are installed
- Add PostgreSQL bin directory to your PATH

**"connection refused"**
- Check that PostgreSQL is running
- Verify PGHOST, PGUSER, and PGDATABASE are set correctly
- Ensure your `.pgpass` file has correct permissions (600)

**"unable to analyze the query"**
- Check your SQL syntax
- Verify you have permissions to run EXPLAIN on the tables
- Ensure the database user has necessary access rights

---

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/Rohatsahin/pgexplain.git
cd pgexplain

# Install dependencies
go mod download

# Build
go build -o pg_explain

# Run tests (when available)
go test ./...
```

---

## Roadmap

Completed features:

- [x] Query comparison mode
- [x] JSON output format
- [x] Cost threshold alerts
- [x] Configuration file support (`.pgexplainrc`)
- [x] Index recommendations

Potential future features:

- [ ] Batch analysis from SQL files
- [ ] Historical plan tracking with SQLite
- [ ] Visual plan diff for comparisons
- [ ] Additional output formats (Markdown, CSV)
- [ ] Query optimization suggestions based on patterns

---

## License

This project is licensed under the Apache 2.0 License. See the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- [Cobra CLI](https://github.com/spf13/cobra) - CLI framework
- [PostgreSQL](https://www.postgresql.org/) - Database system
- [pev2](https://github.com/dalibo/pev2) - Query plan visualization library
- [Dalibo](https://www.dalibo.com/) - Remote explain service

---

## Support

For issues, questions, or contributions:
- GitHub Issues: [https://github.com/Rohatsahin/pgexplain/issues](https://github.com/Rohatsahin/pgexplain/issues)
- Discussions: [https://github.com/Rohatsahin/pgexplain/discussions](https://github.com/Rohatsahin/pgexplain/discussions)

