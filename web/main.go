package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	appCatalog "shopigo/internal/app/catalog"
	infraCatalog "shopigo/internal/infra/catalog"
	infraShared "shopigo/internal/infra/shared"
	httptransport "shopigo/internal/transport/http"
	catalogHTTP "shopigo/internal/transport/http/catalog"
)

const defaultHTTPAddr = ":8089"

func main() {
	err := godotenv.Load() // loads .env into process env
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	startupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	server, repo, err := newServer(startupCtx)
	if err != nil {
		return err
	}
	defer func() { _ = repo.Close() }()

	log.Printf("listening on %s", server.Addr)

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	for {
		select {
		case <-ctx.Done():
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer shutdownCancel()

			if err := server.Shutdown(shutdownCtx); err != nil {
				return err
			}
			return nil
		case err := <-errCh:
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			return err
		}
	}
}

func newServer(ctx context.Context) (*http.Server, *infraCatalog.PostgresCategoryRepository, error) {
	config := infraShared.NewPostgresConfig()

	db, err := config.ConnectContext(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = db.Close() }()

	migrationsDir, err := infraShared.GetMigrationsDir("catalog")
	if err != nil {
		return nil, nil, err
	}

	if err := infraShared.RunMigrations(ctx, db, migrationsDir); err != nil {
		return nil, nil, err
	}

	repo, err := infraCatalog.NewPostgresCategoryRepository(ctx, config.ConnectionString())
	if err != nil {
		return nil, nil, err
	}

	service := appCatalog.NewCategoryService(repo)
	handler := catalogHTTP.NewCategoryHandler(service)

	return &http.Server{
		Addr:              httpAddr(),
		Handler:           httptransport.NewCategoryRouter(handler),
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}, repo, nil
}

func httpAddr() string {
	if addr := os.Getenv("SHOPIGO_HTTP_ADDR"); addr != "" {
		return addr
	}
	return defaultHTTPAddr
}
