package testhelper

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// CreateTestDB creates the test database if it doesn't already exist.
func CreateTestDB(ctx context.Context, adminDB *sqlx.DB, dbName string) error {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)`
	if err := adminDB.GetContext(ctx, &exists, query, dbName); err != nil {
		return fmt.Errorf("failed to check if db exists: %w", err)
	}

	if exists {
		return nil // Database already exists
	}

	createQuery := "CREATE DATABASE " + dbName
	if _, err := adminDB.ExecContext(ctx, createQuery); err != nil {
		return fmt.Errorf("failed to create database %s: %w", dbName, err)
	}

	return nil
}

// DropAllTables drops all public tables in the test database.
func DropAllTables(ctx context.Context, db *sqlx.DB) error {
	query := `
		SELECT tablename FROM pg_tables 
		WHERE schemaname = 'public'
	`
	var tables []string
	if err := db.SelectContext(ctx, &tables, query); err != nil {
		return fmt.Errorf("failed to list tables: %w", err)
	}

	for _, table := range tables {
		dropQuery := "DROP TABLE IF EXISTS " + table + " CASCADE"
		if _, err := db.ExecContext(ctx, dropQuery); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	return nil
}
