// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// UserBalance represents a user's balance in the database
type UserBalance struct {
	UID     string
	Balance int64 // Balance in atoms (1 DCR = 1e11 atoms)
}

// DBManager handles database operations
type DBManager struct {
	db *sql.DB
	mu sync.Mutex
}

// NewDBManager creates a new database manager
func NewDBManager(appRoot string) (*DBManager, error) {
	// Ensure the data directory exists
	dataDir := filepath.Join(appRoot, "data")
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %v", err)
	}

	// Open the database
	dbPath := filepath.Join(dataDir, "balances.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Create the table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_balances (
			uid TEXT PRIMARY KEY,
			balance INTEGER NOT NULL DEFAULT 0
		)
	`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create table: %v", err)
	}

	return &DBManager{
		db: db,
	}, nil
}

// Close closes the database connection
func (dm *DBManager) Close() error {
	return dm.db.Close()
}

// GetBalance retrieves a user's balance
func (dm *DBManager) GetBalance(uid string) (int64, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	var balance int64
	err := dm.db.QueryRow("SELECT balance FROM user_balances WHERE uid = ?", uid).Scan(&balance)
	if err == sql.ErrNoRows {
		// User doesn't exist yet, return 0
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %v", err)
	}
	return balance, nil
}

// UpdateBalance updates a user's balance
func (dm *DBManager) UpdateBalance(uid string, amount int64) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Check if user exists
	var exists bool
	err := dm.db.QueryRow("SELECT EXISTS(SELECT 1 FROM user_balances WHERE uid = ?)", uid).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if user exists: %v", err)
	}

	if exists {
		// Update existing user's balance
		_, err = dm.db.Exec("UPDATE user_balances SET balance = balance + ? WHERE uid = ?", amount, uid)
		if err != nil {
			return fmt.Errorf("failed to update balance: %v", err)
		}
	} else {
		// Insert new user
		_, err = dm.db.Exec("INSERT INTO user_balances (uid, balance) VALUES (?, ?)", uid, amount)
		if err != nil {
			return fmt.Errorf("failed to insert user: %v", err)
		}
	}

	return nil
}
