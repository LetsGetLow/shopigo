package catalog

import (
	"context"
	"database/sql"
	"fmt"
	domain "shopigo/internal/domain/catalog"
	shared "shopigo/internal/domain/shared"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const categoryTable = "catalog_categories"

type categoryRow struct {
	id          uuid.UUID
	parentID    uuid.NullUUID
	name        string
	description string
	createdAt   time.Time
	createdBy   uuid.UUID
	updatedAt   sql.NullTime
	updatedBy   uuid.NullUUID
	deletedAt   sql.NullTime
	deletedBy   uuid.NullUUID
}

func (r categoryRow) toDomain() *domain.Category {
	return &domain.Category{
		ID:          domain.CategoryID(r.id),
		Name:        r.name,
		Description: r.description,
		ParentID:    domain.ParentCategoryIDFromNullUUID(r.parentID),
		Audit: shared.NewAudit(
			shared.CreatedAt(r.createdAt),
			shared.ActorID(r.createdBy),
			shared.UpdatedAtFromNullTime(r.updatedAt),
			shared.ActorIDFromNullUUID(r.updatedBy),
			shared.DeletedAtFromNullTime(r.deletedAt),
			shared.ActorIDFromNullUUID(r.deletedBy),
		),
	}
}

func categoryRowFromDomain(category domain.Category, user shared.ActorID) categoryRow {
	row := categoryRow{
		id:          uuid.UUID(category.ID),
		name:        category.Name,
		description: category.Description,
		createdBy:   uuid.UUID(user),
	}

	if category.ParentID != nil {
		row.parentID = uuid.NullUUID{UUID: uuid.UUID(*category.ParentID), Valid: true}
	}

	return row
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

func (r *PostgresCategoryRepository) Save(ctx context.Context, category domain.Category, user shared.ActorID) error {
	row := categoryRowFromDomain(category, user)
	query := `INSERT INTO ` + categoryTable + ` (category_id, name, description, parent_id, created_at, created_by)
VALUES ($1, $2, $3, $4, NOW(), $5)
ON CONFLICT (category_id) DO UPDATE SET
	name = EXCLUDED.name,
	description = EXCLUDED.description,
	parent_id = EXCLUDED.parent_id,
	updated_at = NOW(),
	updated_by = EXCLUDED.created_by
`

	_, err := r.db.ExecContext(ctx, query,
		row.id,
		row.name,
		row.description,
		row.parentID,
		row.createdBy,
	)
	if err != nil {
		return fmt.Errorf("failed to save category: %w", err)
	}

	return nil
}

func (r *PostgresCategoryRepository) Get(ctx context.Context, id domain.CategoryID) (*domain.Category, error) {
	query := `SELECT category_id, name, description, parent_id, created_at, created_by, updated_at, updated_by, deleted_at, deleted_by
FROM ` + categoryTable + ` WHERE category_id = $1 AND deleted_at IS NULL`
	var row categoryRow
	err := r.db.QueryRowContext(ctx, query, uuid.UUID(id)).Scan(
		&row.id,
		&row.name,
		&row.description,
		&row.parentID,
		&row.createdAt,
		&row.createdBy,
		&row.updatedAt,
		&row.updatedBy,
		&row.deletedAt,
		&row.deletedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return row.toDomain(), nil
}

func (r *PostgresCategoryRepository) List(ctx context.Context) ([]domain.Category, error) {
	return nil, nil
}

func (r *PostgresCategoryRepository) ListByParent(ctx context.Context, id *domain.ParentCategoryID) ([]domain.Category, error) {
	return nil, nil
}

func (r *PostgresCategoryRepository) Delete(ctx context.Context, id domain.CategoryID, user shared.ActorID) error {
	// TODO: add support to delete the whole tree
	query := "UPDATE " + categoryTable + " SET deleted_at = $1, deleted_by = $2 WHERE category_id = $3"
	_, err := r.db.ExecContext(ctx, query, time.Now(), uuid.UUID(user), uuid.UUID(id))
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}
	return nil
}

func (r *PostgresCategoryRepository) Close() error {
	return r.db.Close()
}
