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

	_, err = db.Exec(string(content))
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
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

