package repository

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/denisenkom/go-mssqldb" // MSSQL driver
	_ "github.com/go-sql-driver/mysql"   // MySQL/MariaDB driver
	_ "github.com/jackc/pgx/v5/stdlib"    // Postgres driver
)

// TargetDBConnection contains connection credentials
type TargetDBConnection struct {
	DbType       string
	Host         string
	Port         int
	Username     string
	Password     string
	DatabaseName string
}

// GetDSN returns connection string by database type
func (c *TargetDBConnection) GetDSN() string {
	switch strings.ToLower(c.DbType) {
	case "postgres":
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			c.Username, c.Password, c.Host, c.Port, c.DatabaseName)
	case "mysql", "mariadb":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			c.Username, c.Password, c.Host, c.Port, c.DatabaseName)
	case "mssql":
		return fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;encrypt=disable",
			c.Host, c.Username, c.Password, c.Port, c.DatabaseName)
	default:
		return ""
	}
}

// GetDriverName returns standard driver name
func (c *TargetDBConnection) GetDriverName() string {
	switch strings.ToLower(c.DbType) {
	case "postgres":
		return "pgx"
	case "mysql", "mariadb":
		return "mysql"
	case "mssql":
		return "mssql"
	default:
		return ""
	}
}

// ExecuteExplain runs the DB-specific EXPLAIN query to get execution plan
func ExecuteExplain(conn TargetDBConnection, query string) (string, error) {
	driver := conn.GetDriverName()
	dsn := conn.GetDSN()
	if driver == "" || dsn == "" {
		return "", fmt.Errorf("unsupported database type: %s", conn.DbType)
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return "", fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return "", fmt.Errorf("failed to ping database: %w", err)
	}

	switch strings.ToLower(conn.DbType) {
	case "postgres":
		return runExplainPostgres(db, query)
	case "mysql", "mariadb":
		return runExplainMySQL(db, query)
	case "mssql":
		return runExplainMSSQL(db, query)
	default:
		return "", fmt.Errorf("unsupported explain db: %s", conn.DbType)
	}
}

func runExplainPostgres(db *sql.DB, query string) (string, error) {
	explainQuery := fmt.Sprintf("EXPLAIN %s", query)
	rows, err := db.Query(explainQuery)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var planLines []string
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			return "", err
		}
		planLines = append(planLines, line)
	}
	return strings.Join(planLines, "\n"), nil
}

func runExplainMySQL(db *sql.DB, query string) (string, error) {
	explainQuery := fmt.Sprintf("EXPLAIN FORMAT=JSON %s", query)
	rows, err := db.Query(explainQuery)
	if err != nil {
		// Fallback to simple EXPLAIN if JSON format fails
		explainQuery = fmt.Sprintf("EXPLAIN %s", query)
		rows, err = db.Query(explainQuery)
		if err != nil {
			return "", err
		}
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return "", err
	}

	var results []string
	// For JSON output, it's typically one cell
	if len(cols) == 1 {
		for rows.Next() {
			var val string
			if err := rows.Scan(&val); err != nil {
				return "", err
			}
			results = append(results, val)
		}
		return strings.Join(results, "\n"), nil
	}

	// For tabular explain, build a readable table string
	results = append(results, strings.Join(cols, " | "))
	rawResult := make([][]byte, len(cols))
	dest := make([]interface{}, len(cols))
	for i := range rawResult {
		dest[i] = &rawResult[i]
	}

	for rows.Next() {
		if err := rows.Scan(dest...); err != nil {
			return "", err
		}
		rowVals := make([]string, len(cols))
		for i, raw := range rawResult {
			if raw == nil {
				rowVals[i] = "NULL"
			} else {
				rowVals[i] = string(raw)
			}
		}
		results = append(results, strings.Join(rowVals, " | "))
	}
	return strings.Join(results, "\n"), nil
}

func runExplainMSSQL(db *sql.DB, query string) (string, error) {
	// Enable SHOWPLAN
	_, err := db.Exec("SET SHOWPLAN_ALL ON")
	if err != nil {
		return "", fmt.Errorf("failed to enable showplan: %w", err)
	}
	defer db.Exec("SET SHOWPLAN_ALL OFF")

	rows, err := db.Query(query)
	if err != nil {
		return "", fmt.Errorf("failed to execute explain query: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return "", err
	}

	var results []string
	results = append(results, strings.Join(cols, " | "))

	rawResult := make([][]byte, len(cols))
	dest := make([]interface{}, len(cols))
	for i := range rawResult {
		dest[i] = &rawResult[i]
	}

	for rows.Next() {
		if err := rows.Scan(dest...); err != nil {
			return "", err
		}
		rowVals := make([]string, len(cols))
		for i, raw := range rawResult {
			if raw == nil {
				rowVals[i] = "NULL"
			} else {
				rowVals[i] = string(raw)
			}
		}
		results = append(results, strings.Join(rowVals, " | "))
	}

	return strings.Join(results, "\n"), nil
}

// ExecuteQuery runs a physical query on the target DB (e.g., executing the recommended fix such as CREATE INDEX)
func ExecuteQuery(conn TargetDBConnection, query string) error {
	driver := conn.GetDriverName()
	dsn := conn.GetDSN()
	if driver == "" || dsn == "" {
		return fmt.Errorf("unsupported database type: %s", conn.DbType)
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to execute query fix: %w", err)
	}

	return nil
}
