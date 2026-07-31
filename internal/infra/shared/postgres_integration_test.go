package shared_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	shared "shopigo/internal/infra/shared"
	"shopigo/internal/testhelper"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var testDB *sqlx.DB

const testDatabaseName = "shopigo_shared_test"

func TestMain(m *testing.M) {
	ctx := context.Background()

	if err := os.Setenv("POSTGRES_DB_TEST", testDatabaseName); err != nil {
		fmt.Fprintf(os.Stderr, "failed to set postgres test database: %v\n", err)
		os.Exit(1)
	}

	config := shared.NewPostgresConfigTest()

	adminDB, err := shared.ConnectToAdmin(ctx, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to postgres admin: %v\n", err)
		os.Exit(1)
	}

	if err := testhelper.CreateTestDB(ctx, adminDB, config.DB); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create test db: %v\n", err)
		_ = adminDB.Close()
		os.Exit(1)
	}
	_ = adminDB.Close()

	testDB, err = config.ConnectContext(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to test db: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	if err := testhelper.DropAllTables(ctx, testDB); err != nil {
		fmt.Fprintf(os.Stderr, "failed to cleanup test db: %v\n", err)
	}
	_ = testDB.Close()

	os.Exit(code)
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()

	old, existed := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("failed to unset %s: %v", key, err)
	}

	// revert environment variable
	t.Cleanup(func() {
		if existed {
			if err := os.Setenv(key, old); err != nil {
				t.Errorf("failed to restore %s: %v", key, err)
			}
			return
		}
		_ = os.Unsetenv(key)
	})
}

func chdir(t *testing.T, dir string) {
	t.Helper()

	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change directory to %s: %v", dir, err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(old); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	})
}

func TestPostgresConfig(t *testing.T) {
	t.Run("uses defaults when env is unset", func(t *testing.T) {
		unsetEnv(t, "POSTGRES_HOST")
		unsetEnv(t, "POSTGRES_PORT")
		unsetEnv(t, "POSTGRES_USER")
		unsetEnv(t, "POSTGRES_PASSWORD")
		unsetEnv(t, "POSTGRES_DB")

		cfg := shared.NewPostgresConfig()

		want := shared.PostgresConfig{
			Host:     "localhost",
			Port:     "15432",
			User:     "shopigo",
			Password: "password",
			DB:       "shopigo",
		}

		if cfg != want {
			t.Fatalf("expected config %+v, got %+v", want, cfg)
		}
	})

	t.Run("uses env overrides", func(t *testing.T) {
		t.Setenv("POSTGRES_HOST", "db.internal")
		t.Setenv("POSTGRES_PORT", "5433")
		t.Setenv("POSTGRES_USER", "tester")
		t.Setenv("POSTGRES_PASSWORD", "secret")
		t.Setenv("POSTGRES_DB", "tester")

		cfg := shared.NewPostgresConfig()

		want := shared.PostgresConfig{
			Host:     "db.internal",
			Port:     "5433",
			User:     "tester",
			Password: "secret",
			DB:       "tester",
		}

		if cfg != want {
			t.Fatalf("expected config %+v, got %+v", want, cfg)
		}
	})

	t.Run("uses test db helper when env is unset", func(t *testing.T) {
		unsetEnv(t, "POSTGRES_HOST")
		unsetEnv(t, "POSTGRES_PORT")
		unsetEnv(t, "POSTGRES_USER")
		unsetEnv(t, "POSTGRES_PASSWORD")
		unsetEnv(t, "POSTGRES_DB_TEST")

		cfg := shared.NewPostgresConfigTest()

		want := shared.PostgresConfig{
			Host:     "localhost",
			Port:     "15432",
			User:     "shopigo",
			Password: "password",
			DB:       "shopigo_test",
		}

		if cfg != want {
			t.Fatalf("expected config %+v, got %+v", want, cfg)
		}
	})

	t.Run("formats the connection string", func(t *testing.T) {
		cfg := shared.PostgresConfig{
			Host:     "db.internal",
			Port:     "5432",
			User:     "tester",
			Password: "secret",
			DB:       "shopigo_test",
		}

		got := cfg.ConnectionString()
		want := "postgresql://tester:secret@db.internal:5432/shopigo_test?sslmode=disable"
		if got != want {
			t.Fatalf("expected connection string %q, got %q", want, got)
		}
	})
}

func TestGetMigrationsDir(t *testing.T) {
	t.Run("finds the catalog migrations directory", func(t *testing.T) {
		dir, err := shared.GetMigrationsDir("catalog")
		if err != nil {
			t.Fatalf("expected migrations directory, got error: %v", err)
		}

		if !strings.HasSuffix(filepath.Clean(dir), filepath.Join("migrations", "catalog")) {
			t.Fatalf("expected catalog migrations directory, got %s", dir)
		}

		if _, err := os.Stat(filepath.Join(dir, "001_create_categories.sql")); err != nil {
			t.Fatalf("expected catalog migration file to exist, got: %v", err)
		}
	})

	t.Run("returns file system error when migrations dir is missing", func(t *testing.T) {
		root := t.TempDir()
		chdir(t, root)

		if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example\n"), 0o644); err != nil {
			t.Fatalf("failed to create module file: %v", err)
		}

		_, err := shared.GetMigrationsDir("catalog")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, shared.ErrFileSystem) {
			t.Fatalf("expected ErrFileSystem, got %v", err)
		}
	})

	t.Run("returns not found when go.mod is absent", func(t *testing.T) {
		chdir(t, t.TempDir())

		_, err := shared.GetMigrationsDir("catalog")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, shared.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestRunMigrations(t *testing.T) {
	t.Run("applies sql files in order", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		dir := t.TempDir()
		tableName := "shared_migrations_" + strings.ReplaceAll(uuid.NewString(), "-", "")

		create := fmt.Sprintf(`CREATE TABLE %s (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL
		);`, tableName)
		insert := fmt.Sprintf("INSERT INTO %s (name) VALUES ('first');", tableName)

		if err := os.WriteFile(filepath.Join(dir, "001_create.sql"), []byte(create), 0o644); err != nil {
			t.Fatalf("failed to write create migration: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "002_insert.sql"), []byte(insert), 0o644); err != nil {
			t.Fatalf("failed to write insert migration: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("ignored"), 0o644); err != nil {
			t.Fatalf("failed to write ignored file: %v", err)
		}

		t.Cleanup(func() {
			query := "DROP TABLE IF EXISTS " + tableName
			_, _ = testDB.ExecContext(ctx, query)
		})

		if err := shared.RunMigrations(ctx, testDB, dir); err != nil {
			t.Fatalf("expected migrations to run, got error: %v", err)
		}

		var count int
		if err := testDB.GetContext(ctx, &count, "SELECT COUNT(*) FROM "+tableName); err != nil {
			t.Fatalf("failed to read migrated table: %v", err)
		}
		if count != 1 {
			t.Fatalf("expected 1 row, got %d", count)
		}
	})

	t.Run("wraps sql execution failures", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "001_bad.sql"), []byte("THIS IS NOT SQL;"), 0o644); err != nil {
			t.Fatalf("failed to write invalid migration: %v", err)
		}

		err := shared.RunMigrations(ctx, testDB, dir)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, shared.ErrSqlExecutionFailed) {
			t.Fatalf("expected ErrSqlExecutionFailed, got %v", err)
		}
	})
}

func TestConnectToAdmin(t *testing.T) {
	t.Run("connects to the admin database", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		db, err := shared.ConnectToAdmin(ctx, shared.NewPostgresConfigTest())
		if err != nil {
			t.Fatalf("expected admin connection, got error: %v", err)
		}
		defer db.Close()

		if err := db.PingContext(ctx); err != nil {
			t.Fatalf("expected admin ping to succeed, got: %v", err)
		}
	})

	t.Run("wraps connection failures", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cfg := shared.PostgresConfig{
			Host:     "127.0.0.1",
			Port:     "1",
			User:     "missing",
			Password: "missing",
			DB:       "missing",
		}

		_, err := shared.ConnectToAdmin(ctx, cfg)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var connErr *shared.ConnectionFailed
		if !errors.As(err, &connErr) {
			t.Fatalf("expected ConnectionFailed, got %T", err)
		}
		if connErr.ConnectionName != "postgres admin" {
			t.Fatalf("expected postgres admin connection name, got %q", connErr.ConnectionName)
		}
	})
}

func TestConnectContext(t *testing.T) {
	t.Run("connects to the test database", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		db, err := shared.NewPostgresConfigTest().ConnectContext(ctx)
		if err != nil {
			t.Fatalf("expected test database connection, got error: %v", err)
		}
		defer db.Close()

		if err := db.PingContext(ctx); err != nil {
			t.Fatalf("expected test database ping to succeed, got: %v", err)
		}
	})

	t.Run("wraps connection failures", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cfg := shared.PostgresConfig{
			Host:     "127.0.0.1",
			Port:     "1",
			User:     "missing",
			Password: "missing",
			DB:       "missing",
		}

		_, err := cfg.ConnectContext(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var connErr *shared.ConnectionFailed
		if !errors.As(err, &connErr) {
			t.Fatalf("expected ConnectionFailed, got %T", err)
		}
		if connErr.ConnectionName != "postgres" {
			t.Fatalf("expected postgres connection name, got %q", connErr.ConnectionName)
		}
	})
}
