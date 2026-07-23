package catalog

import (
	"context"
	"fmt"
	domain "shopigo/internal/domain/catalog"
	shared "shopigo/internal/domain/shared"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type categoryRow struct {
	ID          uuid.UUID  `db:"category_id"`
	ParentID    *uuid.UUID `db:"parent_id"`
	Name        string     `db:"name"`
	Description string     `db:"description"`
	CreatedAt   *time.Time `db:"created_at"`
	UpdatedAt   *time.Time `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
	IsDeleted   bool       `db:"is_deleted"`
}

type PostgresCategoryRepository struct {
	db *sqlx.DB
}

func NewPostgresCategoryRepository(ctx context.Context, connectionString string) (*PostgresCategoryRepository, error) {
	db, err := sqlx.ConnectContext(ctx, "postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(0)

	return &PostgresCategoryRepository{db: db}, nil
}

func (r *PostgresCategoryRepository) Save(ctx context.Context, category *domain.Category) error {
	query := `
		INSERT INTO categories (category_id, name, description, parent_id, created_at, is_deleted)
		VALUES ($1, $2, $3, $4, $5, false)
		ON CONFLICT (category_id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			parent_id = EXCLUDED.parent_id,
			updated_at = EXCLUDED.created_at
	`

	_, err := r.db.ExecContext(ctx, query,
		uuid.UUID(category.ID),
		category.Name,
		category.Description,
		(*uuid.UUID)(category.ParentID),
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to save category: %w", err)
	}

	return nil
}

func (r *PostgresCategoryRepository) Get(ctx context.Context, id domain.CategoryID) (*domain.Category, error) {
	query := "SELECT category_id, name, description, parent_id FROM categories WHERE category_id = $1 AND is_deleted = false"
	var row categoryRow
	err := r.db.GetContext(ctx, &row, query, uuid.UUID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	category := &domain.Category{
		ID:          domain.CategoryID(row.ID),
		Name:        row.Name,
		Description: row.Description,
		ParentID:    (*domain.ParentCategoryID)(row.ParentID),
		Audit: shared.NewAudit(
			(*shared.CreatedAt)(row.CreatedAt),
			(*shared.UpdatedAt)(row.UpdatedAt),
			(*shared.DeletedAt)(row.DeletedAt),
		),
	}
	return category, nil
}

func (r *PostgresCategoryRepository) List(ctx context.Context) ([]domain.Category, error) {
	return nil, nil
}

func (r *PostgresCategoryRepository) ListByParent(ctx context.Context, parentID domain.CategoryID) ([]domain.Category, error) {
	return nil, nil
}

func (r *PostgresCategoryRepository) Delete(ctx context.Context, id domain.CategoryID) error {
	return nil
}

func (r *PostgresCategoryRepository) Close() error {
	return r.db.Close()
}
