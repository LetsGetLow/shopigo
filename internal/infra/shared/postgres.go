package shared

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"shopigo/internal/platform/shared"
	"strings"

	"github.com/jmoiron/sqlx"
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
		Host:     shared.LookupConfigValueWithFallback("POSTGRES_HOST", "localhost"),
		Port:     shared.LookupConfigValueWithFallback("POSTGRES_PORT", "15432"),
		User:     shared.LookupConfigValueWithFallback("POSTGRES_USER", "shopigo"),
		Password: shared.LookupConfigValueWithFallback("POSTGRES_PASSWORD", "password"),
		DB:       shared.LookupConfigValueWithFallback("POSTGRES_DB_TEST", "shopigo_test"),
	}
}

// ConnectionString returns the PostgreSQL DSN.
func (c PostgresConfig) ConnectionString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		c.User, c.Password, c.Host, c.Port, c.DB)
}

// ConnectContext opens a connection to the test database.
// It returns a *ConnectionFailed if the connection fails.
func (c PostgresConfig) ConnectContext(ctx context.Context) (*sqlx.DB, error) {
	db, err := sqlx.ConnectContext(ctx, "postgres", c.ConnectionString())
	if err != nil {
		return nil, NewConnectionFailedError("postgres", err)
	}
	return db, nil
}

// GetMigrationsDir finds the migrations directory for a given domain.
// It walks up from the current directory to find go.mod (module root),
// then returns migrations/{domain} relative to it.
func GetMigrationsDir(domain string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", NewOSOperationError(err)
	}

	// TODO: this logic needs to be reworked for production
	current := wd
	for {
		goModPath := filepath.Join(current, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// Found go.mod, so this is the module root
			migrationsDir := filepath.Join(current, "migrations", domain)
			if info, err := os.Stat(migrationsDir); err == nil && info.IsDir() {
				return migrationsDir, nil
			}
			return "", NewFileSystemError(fmt.Errorf("migrations/%s not found at %s", domain, migrationsDir))
		}

		parent := filepath.Dir(current)
		if parent == current {
			// Reached root directory without finding go.mod
			break
		}
		current = parent
	}

	return "", NewNotFoundError("go.mod not found in any parent directory")
}

// RunMigrations executes SQL migration files from the given directory.
// Files are executed in alphabetical order.
func RunMigrations(ctx context.Context, db *sqlx.DB, migrationsDir string) error {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return NewFileSystemError(fmt.Errorf("failed to read migrations dir: %w", err))
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			path := filepath.Join(migrationsDir, entry.Name())
			content, err := os.ReadFile(path)
			if err != nil {
				return NewFileSystemError(fmt.Errorf("failed to read migration %s: %w", entry.Name(), err))
			}

			if _, err := db.ExecContext(ctx, string(content)); err != nil {
				return NewSqlExecutionFailedError(fmt.Errorf("failed to execute migration %s: %w", entry.Name(), err))
			}
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
		return nil, NewConnectionFailedError("postgres admin", err)
	}

	return db, nil
}
