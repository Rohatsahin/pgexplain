<p align="center">
  <img src="./assets/project_image.webp" alt="Project Overview" width="400" height="400" />
  <h3 align="center">PG Explain</h3>
  <p align="center">A command-line tool to analyze and visualize PostgreSQL database queries with using pev2</p>
  <p align="center">
  <a href="https://opensource.org/licenses/Apache-2.0"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="Apache 2.0"></a>
</p>

---

## About The Project

PG Explain is a command-line tool to analyze and visualize PostgreSQL database queries. Built with Go and Cobra, it
provides an intuitive and structured interface for generating and understanding execution plans, with the ability to
visualize them using [pev2](https://github.com/dalibo/pev2).

## Features

- **Command-Oriented**: Designed with Cobra for a structured and user-friendly CLI.
- **Integration with PostgreSQL**: Simplifies database query analysis and visualization.

---

## Installation

you can use `go install` to add the application to your `$GOPATH/bin`:

```bash
  go install github.com/Rohatsahin/pgexplain@latest
```

---

## Configuration

The application relies on PostgreSQL shell environment variables for database connections. Ensure the following
environment variables are set:

- `PGHOST`: PostgreSQL server hostname
- `PGUSER`: PostgreSQL username
- `PGDATABASE`: Target database

### Setting up a `.pgpass` File

You can configure a `.pgpass` file for secure password management:

```bash
  echo "localhost:5432:mydatabase:myuser:mypassword" > ~/.pgpass
  chmod 600 ~/.pgpass
```

For more information on `.pgpass`, see the PostgreSQL
official [documentation](https://www.postgresql.org/docs/current/libpq-pgpass.html).

---

## Build the Application

Once you've set up the app with commands, you'll need to build it into an executable binary.
```bash
  go build -o pg_explain
```

### Move the Binary to a Directory in Your PATH

To make your app accessible globally in the terminal, move the binary file to a directory that is included in your system's PATH environment variable.

On Linux or macOS:
Standard binary location: /usr/local/bin (you might need to use sudo depending on your system's permissions)

Move the binary:
```bash
   sudo mv pg_explain /usr/local/bin/
```

On Windows:
For Windows, you can place the binary in a directory such as C:\Program Files\pg_explain\ and then add that directory to your system’s PATH environment variable.

Copy the binary (pg_explain.exe) to a folder like C:\Program Files\pg_explain\.
Add this folder to your system PATH by:
Right-click This PC → Properties → Advanced system settings → Environment Variables.
Under System variables, select Path → Edit → New, and add C:\Program Files\pg_explain\.

### Verify the Registration
```bash
   pg_explain
```

---

## Usage

### Available Commands

Run the following command to see a list of available commands:

```bash
  pg_explain --help
```

### Example Commands

#### Analyze a SQL Query

```bash
  pg_explain analyze "SELECT * FROM users;"
```

#### Create Execution Plan to a Your Machine

```bash
  pg_explain analyze "SELECT * FROM orders;"
  
  Access the plan from the file system: /.../Plan_Created_on_January_1th_2025_00:00:00.html
```
copying the file path and opening it to your favorite browser

#### Upload Execution Plan to a Remote Server

```bash
  pg_explain analyze --remote "SELECT * FROM orders;"
  
  Access the plan from the remote URL: https://explain.dalibo.com/plan/..
```
copying the remote url and opening it to your favorite browser

### Flags

Use the `--help` flag with any command to see its options. For example:

```bash
  pg_explain analyze --help
```

---

## License

This project is licensed under the Apache 2.0 License. See the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- [Cobra CLI](https://github.com/spf13/cobra)
- [PostgreSQL](https://www.postgresql.org/)
- [pev2](https://github.com/dalibo/pev2)

