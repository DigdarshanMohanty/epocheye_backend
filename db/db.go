package db

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

var Conn *pgx.Conn // Exported variable

func InitDB() {
	dsn := os.Getenv("NEON_DB_URL")
	if dsn == "" {
		log.Fatal("NEON_DB_URL is missing")
	}

	var err error
	Conn, err = pgx.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Test query
	var version string
	if err := Conn.QueryRow(context.Background(), "SELECT version()").Scan(&version); err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	log.Println("Connected to:", version)
}
