package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func InitDB(dbPath string) (*sql.DB, error) {
	dbFolder := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbFolder, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	//Config WAL mode
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		log.Printf("Can't config WAL: %v", err)
	}

	createTableQuery := `CREATE TABLE IF NOT EXISTS focus_sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		start_time INTEGER NOT NULL,
		end_time INTEGER DEFAULT 0,
		status TEXT NOT NULL
	);
	`
	if _, err = db.Exec(createTableQuery); err != nil {
		return nil, err
	}

	return db, nil
}
