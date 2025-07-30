package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/steve/pman/shared/crypto"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaFS embed.FS

type DB struct {
	*sql.DB
}

func Initialize() (*DB, error) {
	dbPath := getDBPath()

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	dbWrapper := &DB{db}

	if err := dbWrapper.createTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	if err := dbWrapper.createDefaultAdmin(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create default admin: %w", err)
	}

	return dbWrapper, nil
}

func getDBPath() string {
	if dbPath := os.Getenv("PMAN_DB_PATH"); dbPath != "" {
		return dbPath
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Warning: Could not get home directory, using current directory for database")
		return "pman.db"
	}

	return filepath.Join(homeDir, ".pman", "pman.db")
}

func (db *DB) createTables() error {
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}

func (db *DB) createDefaultAdmin() error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", "admin@pman.system").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for default admin: %w", err)
	}

	if count > 0 {
		return nil
	}

	hashedPassword, err := crypto.HashPassword("DefaultPassword")
	if err != nil {
		return fmt.Errorf("failed to hash default admin password: %w", err)
	}

	_, err = db.Exec(`
		INSERT INTO users (email, password_hash, role, groups, enabled) 
		VALUES (?, ?, 'admin', 'team1:rw,team2:rw', true)
	`, "admin@pman.system", hashedPassword)

	if err != nil {
		return fmt.Errorf("failed to create default admin: %w", err)
	}

	log.Println("Created default admin user: admin@pman.system / DefaultPassword")
	return nil
}
