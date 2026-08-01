// +build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dbPath := "/opt/linux-hardening-client/data/tokens.db"
	shortToken := "84b9c5f4045d4245e0a9b6ac2227cd8a48d52460cf8087c1"
	refreshToken := "ff1814c30dbc198f718e545ae67389c06aa407814f5a81f4c94eefe04bcbb9e2"
	expiresAt := time.Date(2026, 8, 15, 20, 27, 5, 749826000, time.UTC).Format(time.RFC3339Nano)

	log.Printf("Creating SQLite database at: %s", dbPath)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tokens (
			id INTEGER PRIMARY KEY,
			short_token TEXT NOT NULL,
			refresh_token TEXT NOT NULL,
			expires_at TEXT NOT NULL
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	// Insert tokens
	result, err := db.Exec(`
		INSERT OR REPLACE INTO tokens (id, short_token, refresh_token, expires_at)
		VALUES (1, ?, ?, ?)
	`, shortToken, refreshToken, expiresAt)
	if err != nil {
		log.Fatalf("Failed to insert tokens: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("✅ Tokens saved successfully! Rows affected: %d", rowsAffected)

	// Verify
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM tokens").Scan(&count)
	if err != nil {
		log.Printf("Warning: Could not verify: %v", err)
	} else {
		log.Printf("✅ Verification: %d record(s) in database", count)
	}

	fmt.Println("SUCCESS")
}
