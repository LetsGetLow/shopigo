package testhelper

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// PostgresConfig holds test database connection settings.
type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DB       string
}

// NewPostgresConfig loads config from environment variables.
func NewPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:     getEnv("POSTGRES_HOST", "localhost"),
		Port:     getEnv("POSTGRES_PORT", "15432"),
		User:     getEnv("POSTGRES_USER", "shopigo"),
		Password: getEnv("POSTGRES_PASSWORD", "password"),
		DB:       getEnv("POSTGRES_DB_TEST", "shopigo_test"),
	}
}

// ConnectionString returns the PostgreSQL DSN.
func (c PostgresConfig) ConnectionString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		c.User, c.Password, c.Host, c.Port, c.DB)
}

// ConnectContext opens a connection to the test database.
func (c PostgresConfig) ConnectContext(ctx context.Context) (*sqlx.DB, error) {
	db, err := sqlx.ConnectContext(ctx, "postgres", c.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}
	return db, nil
}

// GetMigrationsDir finds the migrations directory for a given domain.
// It walks up from the current directory to find go.mod (module root),
// then returns migrations/{domain} relative to it.
func GetMigrationsDir(domain string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	current := wd
	for {
		goModPath := filepath.Join(current, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// Found go.mod, so this is the module root
			migrationsDir := filepath.Join(current, "migrations", domain)
			if info, err := os.Stat(migrationsDir); err == nil && info.IsDir() {
				return migrationsDir, nil
			}
			return "", fmt.Errorf("migrations/%s not found at %s", domain, migrationsDir)
		}

		parent := filepath.Dir(current)
		if parent == current {
			// Reached root directory without finding go.mod
			break
		}
		current = parent
	}

	return "", fmt.Errorf("go.mod not found in any parent directory")
}

// RunMigrations executes SQL migration files from the given directory.
// Files are executed in alphabetical order.
func RunMigrations(ctx context.Context, db *sqlx.DB, migrationsDir string) error {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			path := filepath.Join(migrationsDir, entry.Name())
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read migration %s: %w", entry.Name(), err)
			}

			if _, err := db.ExecContext(ctx, string(content)); err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
}

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

	createQuery := fmt.Sprintf("CREATE DATABASE %s", dbName)
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
		if _, err := db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	return nil
}

// ConnectToAdmin returns a connection to the default postgres database
// for administrative operations (like creating test databases).
func ConnectToAdmin(ctx context.Context, config PostgresConfig) (*sqlx.DB, error) {
	adminDSN := fmt.Sprintf("postgresql://%s:%s@%s:%s/postgres?sslmode=disable",
		config.User, config.Password, config.Host, config.Port)

	db, err := sqlx.ConnectContext(ctx, "postgres", adminDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres admin: %w", err)
	}

	return db, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
