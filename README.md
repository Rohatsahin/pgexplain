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
- **Batch Analysis**: Analyze multiple SQL queries from a file with combined or individual reports
- **Query Comparison**: Compare two queries side-by-side to identify the most efficient approach
- **Visual Plan Diff**: Interactive HTML comparison with side-by-side execution plans
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

# Analyze a query - Interactive mode (just paste and press Ctrl+D!)
pg_explain analyze

# Or use editor mode
pg_explain analyze --editor

# Or with traditional string argument
pg_explain analyze "SELECT * FROM users WHERE age > 25"

# Compare two queries interactively
pg_explain compare

# Batch analyze multiple queries from a file
pg_explain batch queries.sql --combined

# Analyze with cost threshold and index recommendations
pg_explain analyze -t 1000 -i

# Generate different output formats
pg_explain analyze --format markdown
pg_explain analyze --format csv

# Use file input for stored queries
pg_explain analyze --file query.sql

# Or pipe from other commands
cat query.sql | pg_explain analyze
```

---

## Flexible Query Input

PG Explain supports multiple ways to provide SQL queries, making it easy to work with any query:

### 1. Interactive Prompt (Default - No Flags Needed!)
**Easiest way** - Just run the command with no arguments, paste your query, and press Ctrl+D:

```bash
# Analyze command - paste your query when prompted
pg_explain analyze

# Compare command - paste two queries when prompted
pg_explain compare
```

**Benefits:**
- No need to escape quotes or special characters
- Perfect for pasting queries from other tools
- Multi-line support built-in
- Clean and intuitive

**How it works:**
```
$ pg_explain analyze

📝 Enter your SQL query (paste or type, press Ctrl+D when done):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SELECT *
FROM users
WHERE age > 25
[Press Ctrl+D]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ Query received!
```

### 2. Editor Mode (`--editor` or `-e`)
Opens your favorite editor (vim, nano, VS Code, etc.) to write/paste queries:

```bash
# Analyze in editor
pg_explain analyze --editor

# Compare in editor (opens twice - once for each query)
pg_explain compare --editor
```

**Benefits:**
- Full editor features (syntax highlighting, autocomplete, etc.)
- Uses your $EDITOR environment variable
- Great for complex queries
- Edit comfortably, save and close to analyze

**How it works:**
```
$ pg_explain analyze --editor

✏️  Opening editor: vim
💡 Write your query, save, and close the editor to continue...

[Your editor opens with a .sql file - write your query, :wq to save and exit]

🔍 Analyzing your query...
```

### 3. File Input (`--file` flag)
Load queries from existing SQL files:

```bash
# Analyze from file
pg_explain analyze --file my_complex_query.sql

# Compare from files
pg_explain compare --file1 query1.sql --file2 query2.sql
```

**Benefits:**
- Store queries in version control
- Reuse queries easily
- Share queries with team members
- Keep a library of test queries

### 4. STDIN/Pipe Support
Pipe queries from files or other commands:

```bash
# From a file
cat query.sql | pg_explain analyze

# From echo
echo "SELECT * FROM users" | pg_explain analyze

# From other commands
psql -c "\d users" | grep "Column" | pg_explain analyze
```

**Benefits:**
- Integration with shell scripts and pipelines
- Process queries from other tools
- Automation and scripting

### 5. Command Argument (Still Supported)
Traditional method with query as string argument:

```bash
pg_explain analyze "SELECT * FROM users WHERE age > 25"
pg_explain compare "SELECT * FROM users" "SELECT * FROM orders"
```

**Note:** String arguments are still supported but **interactive modes are recommended** for better usability!

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
pg_explain analyze [SQL_QUERY] [flags]
```

**Query Input Methods (in order of priority):**
1. Interactive prompt: `pg_explain analyze` (default - just press enter and paste!)
2. Editor mode: `pg_explain analyze --editor`
3. File input: `pg_explain analyze --file query.sql`
4. STDIN/Pipe: `cat query.sql | pg_explain analyze`
5. Command argument: `pg_explain analyze "SELECT..."` (still supported)

**Available Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--editor` | `-e` | bool | `false` | Open $EDITOR to write/paste query |
| `--file` | `-F` | string | `""` | Read SQL query from file |
| `--format` | `-f` | string | `html` | Output format: `html`, `json`, `markdown`, or `csv` |
| `--remote` | `-r` | bool | `false` | Upload plan to remote server for sharing |
| `--threshold` | `-t` | float | `0` | Cost threshold for alerting (0 = disabled) |
| `--recommend-indexes` | `-i` | bool | `false` | Recommend indexes based on query execution plan |
| `--index-threshold` | | float | `100` | Minimum operation cost to trigger index recommendations |

---

#### `compare` - Compare two SQL queries

```bash
pg_explain compare [QUERY1] [QUERY2] [flags]
```

**Query Input Methods (in order of priority):**
1. Interactive prompt: `pg_explain compare` (default - paste both queries when prompted!)
2. Editor mode: `pg_explain compare --editor` (opens editor twice)
3. File inputs: `pg_explain compare --file1 q1.sql --file2 q2.sql`
4. Command arguments: `pg_explain compare "SELECT..." "SELECT..."` (still supported)
5. Mixed: `pg_explain compare --file1 q1.sql "SELECT..."`

**Available Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--editor` | `-e` | bool | `false` | Open $EDITOR to write/paste queries (opens twice) |
| `--file1` | | string | `""` | Read first SQL query from file |
| `--file2` | | string | `""` | Read second SQL query from file |
| `--format` | `-f` | string | `text` | Output format: `text`, `json`, `html`, `markdown`, or `csv` |

**Output Formats:**
- `text`: Terminal-based comparison (default)
- `json`: Machine-readable JSON format
- `html`: Interactive visual diff with side-by-side comparison
- `markdown`: Rich formatted markdown with tables and code blocks
- `csv`: Comma-separated values for spreadsheet analysis

---

#### `batch` - Batch analyze SQL queries from a file

```bash
pg_explain batch [SQL_FILE] [flags]
```

**Available Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--format` | `-f` | string | `html` | Output format: `html`, `json`, `markdown`, or `csv` |
| `--threshold` | `-t` | float | `0` | Cost threshold for alerting (0 = disabled) |
| `--recommend-indexes` | `-i` | bool | `false` | Recommend indexes based on query execution plans |
| `--index-threshold` | | float | `100` | Minimum operation cost to trigger index recommendations |
| `--combined` | `-c` | bool | `false` | Generate a single combined report instead of individual files |
| `--output-dir` | `-o` | string | `""` | Directory to save output files (default: current directory) |
| `--continue-on-error` | | bool | `true` | Continue processing remaining queries if one fails |

**SQL File Format:**

Queries should be separated by semicolons (`;`). Empty lines and SQL comments (`--`) are automatically ignored.

**Example SQL File (queries.sql):**
```sql
-- Query 1: Get active users
SELECT * FROM users WHERE status = 'active';

-- Query 2: Get recent orders
SELECT * FROM orders WHERE created_at > NOW() - INTERVAL '7 days';

-- Query 3: Join users and orders
SELECT u.name, COUNT(o.id) as order_count
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
GROUP BY u.name;
```

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

#### 9. Batch Analysis - Individual Files

Analyze multiple queries from a SQL file and generate individual HTML reports for each:

```bash
pg_explain batch queries.sql
```

**Terminal Output:**
```
🔍 Starting batch analysis...
📁 SQL file: queries.sql
📊 Output format: html
📦 Mode: Individual files

✅ Found 3 queries to analyze

🔄 Processing query 1/3...
   ✅ Query 1 analyzed successfully

🔄 Processing query 2/3...
   ✅ Query 2 analyzed successfully

🔄 Processing query 3/3...
   ✅ Query 3 analyzed successfully

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 Batch Analysis Complete
   Total: 3 | Success: 3 | Failed: 0
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💾 Generating individual files...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📁 Generated 3 files successfully!
   /path/to/Query_1_Plan_Created_on_January_11th_2026_10:30:00.html
   /path/to/Query_2_Plan_Created_on_January_11th_2026_10:30:01.html
   /path/to/Query_3_Plan_Created_on_January_11th_2026_10:30:02.html
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

#### 10. Batch Analysis - Combined Report

Generate a single combined HTML report with all queries:

```bash
pg_explain batch queries.sql --combined
```

**Terminal Output:**
```
🔍 Starting batch analysis...
📁 SQL file: queries.sql
📊 Output format: html
📦 Mode: Combined report

✅ Found 3 queries to analyze

🔄 Processing query 1/3...
   ✅ Query 1 analyzed successfully

🔄 Processing query 2/3...
   ✅ Query 2 analyzed successfully

🔄 Processing query 3/3...
   ✅ Query 3 analyzed successfully

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 Batch Analysis Complete
   Total: 3 | Success: 3 | Failed: 0
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💾 Generating combined report...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📁 Batch report saved successfully!
   /path/to/Batch_queries_2026-01-11_10-30-00.html

💡 Tip: Open this file in your browser to view all query plans
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

The combined HTML report features:
- Interactive dashboard with success/failure statistics
- Expandable cards for each query
- Side-by-side query and execution plan view
- Cost analysis for each query
- Easy navigation between queries

---

#### 11. Batch Analysis with Cost Threshold

Analyze multiple queries with cost threshold warnings:

```bash
pg_explain batch queries.sql --threshold 1000 --combined
```

**Terminal Output:**
```
🔍 Starting batch analysis...
📁 SQL file: queries.sql
📊 Output format: html
⚡ Cost threshold: 1000
📦 Mode: Combined report

✅ Found 3 queries to analyze

🔄 Processing query 1/3...
   ✅ Query 1 cost: 425.50

🔄 Processing query 2/3...
   ⚠️  Query 2 exceeds cost threshold (1250.00 > 1000)

🔄 Processing query 3/3...
   ✅ Query 3 cost: 789.20

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 Batch Analysis Complete
   Total: 3 | Success: 3 | Failed: 0
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

#### 12. Batch Analysis - JSON with Custom Output Directory

Save batch results as JSON files in a specific directory:

```bash
pg_explain batch queries.sql --format json --output-dir ./batch-results
```

**Terminal Output:**
```
🔍 Starting batch analysis...
📁 SQL file: queries.sql
📊 Output format: json
📦 Mode: Individual files

✅ Found 3 queries to analyze

🔄 Processing query 1/3...
   ✅ Query 1 analyzed successfully

🔄 Processing query 2/3...
   ✅ Query 2 analyzed successfully

🔄 Processing query 3/3...
   ✅ Query 3 analyzed successfully

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 Batch Analysis Complete
   Total: 3 | Success: 3 | Failed: 0
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💾 Generating individual files...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📁 Generated 3 files successfully!
   /path/to/batch-results/Query_1_Plan_Created_on_January_11th_2026_10:30:00.json
   ... and 2 more files

💡 All files saved to: ./batch-results
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

#### 13. Batch Analysis - Combined JSON Report

Generate a single JSON file with all batch analysis results:

```bash
pg_explain batch queries.sql --format json --combined --threshold 500
```

**JSON Output (Batch_queries_2026-01-11_10-30-00.json):**
```json
{
  "file_name": "queries.sql",
  "total_queries": 3,
  "success_count": 3,
  "failure_count": 0,
  "results": [
    {
      "query_number": 1,
      "query": "SELECT * FROM users WHERE status = 'active'",
      "execution_plan": "Seq Scan on users  (cost=0.00..425.50 rows=1000 width=244)...",
      "cost_analysis": {
        "TotalCost": 425.50,
        "ExpensiveOps": [],
        "ExceedsLimit": false,
        "ThresholdValue": 500
      },
      "generated_at": "2026-01-11T10:30:00Z"
    },
    {
      "query_number": 2,
      "query": "SELECT * FROM orders WHERE created_at > NOW() - INTERVAL '7 days'",
      "execution_plan": "Seq Scan on orders  (cost=0.00..1250.00 rows=5000 width=100)...",
      "cost_analysis": {
        "TotalCost": 1250.00,
        "ExpensiveOps": [
          {
            "Operation": "Seq Scan",
            "Cost": 1250.00,
            "Line": "Seq Scan on orders  (cost=0.00..1250.00 rows=5000 width=100)"
          }
        ],
        "ExceedsLimit": true,
        "ThresholdValue": 500
      },
      "generated_at": "2026-01-11T10:30:01Z"
    },
    {
      "query_number": 3,
      "query": "SELECT u.name, COUNT(o.id) as order_count FROM users u LEFT JOIN orders o ON u.id = o.user_id GROUP BY u.name",
      "execution_plan": "HashAggregate  (cost=789.20..850.25 rows=1000 width=50)...",
      "cost_analysis": {
        "TotalCost": 850.25,
        "ExpensiveOps": [
          {
            "Operation": "HashAggregate",
            "Cost": 850.25,
            "Line": "HashAggregate  (cost=789.20..850.25 rows=1000 width=50)"
          }
        ],
        "ExceedsLimit": true,
        "ThresholdValue": 500
      },
      "generated_at": "2026-01-11T10:30:02Z"
    }
  ],
  "generated_at": "2026-01-11T10:30:02Z"
}
```

This JSON format is perfect for:
- Automated testing and CI/CD pipelines
- Programmatic analysis of query performance
- Tracking query performance over time
- Integration with monitoring tools

---

#### 14. Batch Analysis with Index Recommendations

Get index recommendations for all queries in a file:

```bash
pg_explain batch queries.sql --recommend-indexes --combined
```

**Terminal Output:**
```
🔍 Starting batch analysis...
📁 SQL file: queries.sql
📊 Output format: html
📦 Mode: Combined report

✅ Found 3 queries to analyze

🔄 Processing query 1/3...
   ✅ Query 1 analyzed successfully
   💡 Found 2 index recommendations

🔄 Processing query 2/3...
   ✅ Query 2 analyzed successfully

🔄 Processing query 3/3...
   ✅ Query 3 analyzed successfully
   💡 Found 1 index recommendations

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 Batch Analysis Complete
   Total: 3 | Success: 3 | Failed: 0
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

#### 15. Visual Plan Diff - Interactive HTML Comparison

Generate an interactive visual comparison with side-by-side execution plans:

```bash
pg_explain compare -f html "SELECT * FROM users WHERE created_at > '2024-01-01'" "SELECT * FROM users WHERE created_at >= '2024-01-01' AND created_at < '2025-01-01'"
```

**Terminal Output:**
```
🔬 Starting query comparison...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔍 Analyzing Query 1...
✅ Query 1 complete!

🔍 Analyzing Query 2...
✅ Query 2 complete!

💾 Generating visual comparison report...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📁 Visual comparison report saved successfully!
   /path/to/Comparison_Plan_Created_on_January_11th_2026_10:45:00.html

🏆 Winner: Query 2
Cost Difference: 45.25 (12.50%)

💡 Tip: Open this file in your browser to view the interactive visual diff
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**Visual Diff Features:**
The interactive HTML report includes:
- **Winner Badge**: Prominently displays which query performed better
- **Cost Visualization**: Animated bar chart showing cost comparison
- **Performance Stats**: Cost difference, percentage, and performance multiplier
- **Side-by-Side Comparison**: Query SQL and execution plans displayed adjacently
- **Expensive Operations**: Highlighted operations above cost thresholds
- **Responsive Design**: Beautiful gradient backgrounds and hover effects
- **Easy Navigation**: Interactive panels for each query

**When to Use:**
- Comparing different query syntax approaches
- Evaluating optimization attempts
- Sharing comparison results with team members
- Presenting query performance in meetings
- Documentation and performance audits

---

#### 16. Markdown Format - Rich Formatted Output

Generate Markdown files for documentation and version control:

```bash
pg_explain analyze -f markdown "SELECT * FROM users WHERE status = 'active'"
```

**Terminal Output:**
```
🔍 Analyzing your query...
📊 Output format: markdown

✅ Query analysis complete!

💾 Generating Markdown report...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📁 Plan saved successfully!
   /path/to/Plan_Created_on_January_11th_2026_16:15:00.md
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**Markdown File Contents:**
```markdown
# Query Execution Plan

**Generated:** January 11, 2026 16:15:00
**Query:** SELECT * FROM users WHERE status = 'active'

---

## Cost Analysis

| Metric | Value |
|--------|-------|
| Total Cost | 425.50 |
| Exceeds Threshold | false |
| Threshold Value | 1000.00 |

### Expensive Operations

| Operation | Cost | Details |
|-----------|------|---------|
| Seq Scan | 425.50 | Full table scan on users |

---

## Execution Plan

```
Seq Scan on users  (cost=0.00..425.50 rows=1000 width=244)
  Filter: (status = 'active'::text)
```

---

**Note:** This plan was generated using PostgreSQL EXPLAIN
```

**When to Use Markdown:**
- Documentation in Git repositories
- Pull request descriptions
- Technical reports
- Team wikis and knowledge bases
- GitHub/GitLab issue descriptions

---

#### 17. CSV Format - Spreadsheet Analysis

Generate CSV files for data analysis in Excel or Google Sheets:

```bash
pg_explain analyze -f csv "SELECT * FROM orders WHERE total > 100"
```

**Terminal Output:**
```
🔍 Analyzing your query...
📊 Output format: csv

✅ Query analysis complete!

💾 Saving as CSV...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📁 Plan saved successfully!
   /path/to/Plan_Created_on_January_11th_2026_16:20:00.csv
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**CSV File Structure:**
```csv
"title","query","execution_plan","total_cost","exceeds_threshold","threshold_value","expensive_ops_count","generated_at"
"Plan_Created_on_January_11th_2026_16:20:00","SELECT * FROM orders WHERE total > 100","Seq Scan on orders  (cost=0.00..1250.00 rows=5000 width=100)\n  Filter: (total > 100)","1250.00","false","0.00","1","2026-01-11T16:20:00Z"
```

**When to Use CSV:**
- Bulk query performance analysis
- Tracking query costs over time
- Integration with data analysis tools
- Automated performance monitoring
- Spreadsheet-based reporting

---

#### 18. Compare with Markdown Format

Generate Markdown comparison reports for documentation:

```bash
pg_explain compare -f markdown "SELECT * FROM users WHERE id = 1" "SELECT * FROM users WHERE user_id = 1"
```

**File Features:**
- Winner badge with emoji
- Cost comparison tables
- Side-by-side query and execution plan sections
- Performance multiplier calculations
- Detailed comparison metrics

---

#### 19. Batch Analysis with CSV Format

Analyze multiple queries and export to CSV for spreadsheet analysis:

```bash
pg_explain batch queries.sql --format csv --combined
```

**CSV Output (One Row Per Query):**
```csv
"query_number","query","execution_plan","total_cost","exceeds_threshold","error","status","generated_at"
"1","SELECT * FROM users","Seq Scan on users...\n...","425.50","false","","success","2026-01-11T16:25:00Z"
"2","SELECT * FROM orders","Index Scan...\n...","245.67","false","","success","2026-01-11T16:25:01Z"
"3","SELECT * FROM invalid","","","","relation ""invalid"" does not exist","failed","2026-01-11T16:25:02Z"
```

**Use Cases:**
- Performance trend analysis across multiple queries
- Bulk query auditing
- CI/CD performance regression testing
- Database optimization reports
- Historical performance tracking

---

#### 20. Batch Analysis with Markdown Format

Generate comprehensive Markdown reports for batch analysis:

```bash
pg_explain batch queries.sql --format markdown --combined --threshold 500
```

**Markdown Report Features:**
- Summary statistics table
- Individual query sections with status badges (✅/❌)
- Cost analysis for each query
- Execution plans in code blocks
- Performance summary table at the end

**Perfect For:**
- Team performance reviews
- Database optimization documentation
- Git repository documentation
- Performance audit reports

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
- [x] Batch analysis from SQL files
- [x] Visual plan diff for comparisons
- [x] Additional output formats (Markdown, CSV)

Potential future features:

- [ ] Historical plan tracking with SQLite
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

