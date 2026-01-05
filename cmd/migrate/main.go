package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run cmd/migrate/main.go [up|down]")
	}

	command := os.Args[1]

	// Connect to database
	dsn := "voucheruser:voucherpass@tcp(localhost:3306)/voucher_db?parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	switch command {
	case "up":
		log.Println("Running migrations...")
		if err := runMigrations(db); err != nil {
			log.Fatal("Migration failed:", err)
		}
		log.Println("Migrations completed successfully")
	case "down":
		log.Println("Rolling back migrations...")
		if err := rollbackMigrations(db); err != nil {
			log.Fatal("Rollback failed:", err)
		}
		log.Println("Rollback completed successfully")
	default:
		log.Fatal("Unknown command:", command)
	}
}

func runMigrations(db *sql.DB) error {
	// Read and execute the migration file
	content, err := os.ReadFile("migrations/init.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Split SQL by semicolons and execute each statement
	statements := splitSQL(string(content))
	for i, stmt := range statements {
		stmt = trimStatement(stmt)
		if stmt == "" {
			continue
		}
		
		// Skip USE statements as we're already connected to the database
		if len(stmt) >= 3 && stmt[:3] == "USE" {
			continue
		}
		
		_, err = db.Exec(stmt)
		if err != nil {
			return fmt.Errorf("failed to execute statement %d: %w\nStatement: %s", i+1, err, stmt)
		}
	}

	return nil
}

// splitSQL splits SQL content by semicolons (simple version)
func splitSQL(content string) []string {
	var statements []string
	var current string
	
	for _, line := range splitLines(content) {
		// Skip comments
		line = trimStatement(line)
		if len(line) >= 2 && line[:2] == "--" {
			continue
		}
		
		current += line + "\n"
		
		// Check if statement ends with semicolon
		if len(line) > 0 && line[len(line)-1] == ';' {
			statements = append(statements, current)
			current = ""
		}
	}
	
	if trimStatement(current) != "" {
		statements = append(statements, current)
	}
	
	return statements
}

// splitLines splits content by newlines
func splitLines(content string) []string {
	var lines []string
	var current string
	
	for _, c := range content {
		if c == '\n' || c == '\r' {
			if current != "" || len(lines) == 0 {
				lines = append(lines, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	
	if current != "" {
		lines = append(lines, current)
	}
	
	return lines
}

// trimStatement trims whitespace from a statement
func trimStatement(stmt string) string {
	// Trim leading and trailing whitespace
	start := 0
	end := len(stmt)
	
	for start < end && (stmt[start] == ' ' || stmt[start] == '\t' || stmt[start] == '\n' || stmt[start] == '\r') {
		start++
	}
	
	for end > start && (stmt[end-1] == ' ' || stmt[end-1] == '\t' || stmt[end-1] == '\n' || stmt[end-1] == '\r') {
		end--
	}
	
	return stmt[start:end]
}

func rollbackMigrations(db *sql.DB) error {
	// Drop all tables
	tables := []string{
		"payment_events",
		"purchased_vouchers",
		"transactions",
		"vouchers",
		"wallets",
		"users",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table))
		if err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
		log.Printf("Dropped table: %s", table)
	}

	return nil
}

