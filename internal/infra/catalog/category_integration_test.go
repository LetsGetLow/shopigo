package catalog_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	shared2 "shopigo/internal/infra/shared"
	"testing"
	"time"

	"shopigo/internal/domain/catalog"
	shared "shopigo/internal/domain/shared"
	catrepo "shopigo/internal/infra/catalog"
	"shopigo/internal/testhelper"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var testDB *sqlx.DB

// TODO: write better tests
type categoryAuditRow struct {
	createdAt time.Time
	createdBy uuid.UUID
	updatedAt sql.NullTime
	updatedBy uuid.NullUUID
	deletedAt sql.NullTime
	deletedBy uuid.NullUUID
}

func assertCategoryAuditState(t *testing.T, got *catalog.Category, wantCreatedBy shared.ActorID, wantUpdatedBy *shared.ActorID, wantDeletedBy *shared.ActorID) {
	t.Helper()

	if got == nil {
		t.Fatal("expected category, got nil")
	}

	if time.Time(got.CreatedAt()).IsZero() {
		t.Fatal("expected created at to be set")
	}
	if got.CreatedBy() != wantCreatedBy {
		t.Fatalf("expected created by %v, got %v", wantCreatedBy, got.CreatedBy())
	}

	if wantUpdatedBy == nil {
		if got.UpdatedAt() != nil {
			t.Fatalf("expected updated at to be nil, got %v", got.UpdatedAt())
		}
		if got.UpdatedBy() != nil {
			t.Fatalf("expected updated by to be nil, got %v", got.UpdatedBy())
		}
	} else {
		if got.UpdatedAt() == nil {
			t.Fatal("expected updated at to be set")
		}
		if got.UpdatedBy() == nil {
			t.Fatal("expected updated by to be set")
		}
		if *got.UpdatedBy() != *wantUpdatedBy {
			t.Fatalf("expected updated by %v, got %v", *wantUpdatedBy, *got.UpdatedBy())
		}
	}

	if wantDeletedBy == nil {
		if got.DeletedAt() != nil {
			t.Fatalf("expected deleted at to be nil, got %v", got.DeletedAt())
		}
		if got.DeletedBy() != nil {
			t.Fatalf("expected deleted by to be nil, got %v", got.DeletedBy())
		}
	} else {
		if got.DeletedAt() == nil {
			t.Fatal("expected deleted at to be set")
		}
		if got.DeletedBy() == nil {
			t.Fatal("expected deleted by to be set")
		}
		if *got.DeletedBy() != *wantDeletedBy {
			t.Fatalf("expected deleted by %v, got %v", *wantDeletedBy, *got.DeletedBy())
		}
	}
}

func loadCategoryAuditRow(t *testing.T, id catalog.CategoryID) categoryAuditRow {
	t.Helper()

	var row categoryAuditRow
	err := testDB.QueryRow(
		`SELECT created_at, created_by, updated_at, updated_by, deleted_at, deleted_by
		 FROM catalog_categories
		 WHERE category_id = $1`,
		uuid.UUID(id),
	).Scan(
		&row.createdAt,
		&row.createdBy,
		&row.updatedAt,
		&row.updatedBy,
		&row.deletedAt,
		&row.deletedBy,
	)
	if err != nil {
		t.Fatalf("failed to load category audit row: %v", err)
	}

	return row
}

// TestMain sets up the test database before running integration tests.
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Load test database config
	config := shared2.NewPostgresConfig()

	// Connect to admin database to create test DB
	adminDB, err := shared2.ConnectToAdmin(ctx, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to postgres admin: %v\n", err)
		os.Exit(1)
	}

	// Create test database if it doesn't exist
	if err := testhelper.CreateTestDB(ctx, adminDB, config.DB); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create test db: %v\n", err)
		adminDB.Close()
		os.Exit(1)
	}
	adminDB.Close()

	// Connect to test database
	testDB, err = config.ConnectContext(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to test db: %v\n", err)
		os.Exit(1)
	}
	defer testDB.Close()

	// Get migrations directory for catalog domain
	migrationsDir, err := shared2.GetMigrationsDir("catalog")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to find migrations: %v\n", err)
		os.Exit(1)
	}

	// Run migrations
	if err := shared2.RunMigrations(ctx, testDB, migrationsDir); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	if err := testhelper.DropAllTables(ctx, testDB); err != nil {
		fmt.Fprintf(os.Stderr, "failed to cleanup: %v\n", err)
	}

	os.Exit(code)
}

// TestSaveCategory tests saving a new category and updating an existing one.
func TestSaveCategory(t *testing.T) {
	ctx := context.Background()
	repo, err := catrepo.NewPostgresCategoryRepository(ctx, shared2.NewPostgresConfig().ConnectionString())
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}
	defer repo.Close()

	// Test data
	userID := shared.ActorID(uuid.New())

	// Create a new category
	cat := catalog.NewCategory("Electronics", "Electronics and gadgets")

	// Save the category
	if err := repo.Save(ctx, *cat, userID); err != nil {
		t.Fatalf("failed to save category: %v", err)
	}

	// Verify it was saved by retrieving it
	retrieved, err := repo.Get(ctx, cat.ID)
	if err != nil {
		t.Fatalf("failed to get category: %v", err)
	}

	// Verify the data matches
	if retrieved.Name != cat.Name {
		t.Errorf("expected name %q, got %q", cat.Name, retrieved.Name)
	}
	if retrieved.Description != cat.Description {
		t.Errorf("expected description %q, got %q", cat.Description, retrieved.Description)
	}
	if retrieved.ParentID != cat.ParentID {
		t.Errorf("expected parent ID %v, got %v", cat.ParentID, retrieved.ParentID)
	}
	assertCategoryAuditState(t, retrieved, userID, nil, nil)
	createdAt := retrieved.CreatedAt()

	// Test update
	retrieved.Name = "Updated Electronics"
	if err := repo.Save(ctx, *retrieved, userID); err != nil {
		t.Fatalf("failed to update category: %v", err)
	}

	// Verify the update
	updated, err := repo.Get(ctx, retrieved.ID)
	if err != nil {
		t.Fatalf("failed to get updated category: %v", err)
	}

	if updated.Name != "Updated Electronics" {
		t.Errorf("expected updated name %q, got %q", "Updated Electronics", updated.Name)
	}
	if updated.CreatedAt() != createdAt {
		t.Fatalf("expected created at %v, got %v", createdAt, updated.CreatedAt())
	}
	assertCategoryAuditState(t, updated, userID, &userID, nil)
}

// TestGetCategory tests retrieving a category by ID.
func TestGetCategory(t *testing.T) {
	ctx := context.Background()
	repo, err := catrepo.NewPostgresCategoryRepository(ctx, shared2.NewPostgresConfig().ConnectionString())
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}
	defer repo.Close()

	userID := shared.ActorID(uuid.New())

	// Create and save a category
	cat := catalog.NewCategory("Books", "Books and literature")
	if err := repo.Save(ctx, *cat, userID); err != nil {
		t.Fatalf("failed to save category: %v", err)
	}

	// Retrieve it
	retrieved, err := repo.Get(ctx, cat.ID)
	if err != nil {
		t.Fatalf("failed to get category: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected category, got nil")
	}

	if retrieved.ID != cat.ID {
		t.Errorf("expected ID %v, got %v", cat.ID, retrieved.ID)
	}
	if retrieved.Name != cat.Name {
		t.Errorf("expected name %q, got %q", cat.Name, retrieved.Name)
	}
	assertCategoryAuditState(t, retrieved, userID, nil, nil)

	// Test non-existent category
	nonExistentID := catalog.CategoryID(uuid.New())
	_, err = repo.Get(ctx, nonExistentID)
	if err == nil {
		t.Fatal("expected error for non-existent category, got nil")
	}
}

// TestGetDeletedCategory tests that deleted categories are not retrieved.
func TestGetDeletedCategory(t *testing.T) {
	ctx := context.Background()
	repo, err := catrepo.NewPostgresCategoryRepository(ctx, shared2.NewPostgresConfig().ConnectionString())
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}
	defer repo.Close()

	userID := shared.ActorID(uuid.New())

	// Create and save a category
	cat := catalog.NewCategory("Toys", "Toys and games")
	if err := repo.Save(ctx, *cat, userID); err != nil {
		t.Fatalf("failed to save category: %v", err)
	}

	// Delete it
	if err := repo.Delete(ctx, cat.ID, userID); err != nil {
		t.Fatalf("failed to delete category: %v", err)
	}

	row := loadCategoryAuditRow(t, cat.ID)
	if !row.deletedAt.Valid {
		t.Fatal("expected deleted at to be set")
	}
	if !row.deletedBy.Valid {
		t.Fatal("expected deleted by to be set")
	}
	if row.deletedBy.UUID != uuid.UUID(userID) {
		t.Fatalf("expected deleted by %v, got %v", userID, row.deletedBy.UUID)
	}

	// Try to retrieve it - should fail
	_, err = repo.Get(ctx, cat.ID)
	if err == nil {
		t.Fatal("expected error for deleted category, got nil")
	}
}

// TestDeleteCategory tests soft-deleting a category.
func TestDeleteCategory(t *testing.T) {
	ctx := context.Background()
	repo, err := catrepo.NewPostgresCategoryRepository(ctx, shared2.NewPostgresConfig().ConnectionString())
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}
	defer repo.Close()

	userID := shared.ActorID(uuid.New())

	// Create and save a category
	cat := catalog.NewCategory("Sports", "Sports and fitness")
	if err := repo.Save(ctx, *cat, userID); err != nil {
		t.Fatalf("failed to save category: %v", err)
	}

	// Verify it exists
	_, err = repo.Get(ctx, cat.ID)
	if err != nil {
		t.Fatalf("failed to get category before delete: %v", err)
	}

	// Delete it
	if err := repo.Delete(ctx, cat.ID, userID); err != nil {
		t.Fatalf("failed to delete category: %v", err)
	}

	row := loadCategoryAuditRow(t, cat.ID)
	if !row.deletedAt.Valid {
		t.Fatal("expected deleted at to be set")
	}
	if !row.deletedBy.Valid {
		t.Fatal("expected deleted by to be set")
	}
	if row.deletedBy.UUID != uuid.UUID(userID) {
		t.Fatalf("expected deleted by %v, got %v", userID, row.deletedBy.UUID)
	}

	// Verify it's deleted by checking Get returns error
	_, err = repo.Get(ctx, cat.ID)
	if err == nil {
		t.Fatal("expected error after deletion, got nil")
	}
}

// TestCategoryWithParent tests saving and retrieving a category with a parent.
func TestCategoryWithParent(t *testing.T) {
	ctx := context.Background()
	repo, err := catrepo.NewPostgresCategoryRepository(ctx, shared2.NewPostgresConfig().ConnectionString())
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}
	defer repo.Close()

	userID := shared.ActorID(uuid.New())

	// Create and save parent category
	parent := catalog.NewCategory("Computers", "Computers and peripherals")
	if err := repo.Save(ctx, *parent, userID); err != nil {
		t.Fatalf("failed to save parent category: %v", err)
	}

	// Create child category
	child := catalog.NewCategory("Laptops", "Laptop computers")
	parentID := catalog.ParentCategoryID(parent.ID)
	child.MoveToParent(parentID)

	// Save child category
	if err := repo.Save(ctx, *child, userID); err != nil {
		t.Fatalf("failed to save child category: %v", err)
	}

	// Retrieve child and verify parent relationship
	retrieved, err := repo.Get(ctx, child.ID)
	if err != nil {
		t.Fatalf("failed to get child category: %v", err)
	}

	if retrieved.ParentID == nil {
		t.Fatal("expected parent ID, got nil")
	}
	if *retrieved.ParentID != parentID {
		t.Errorf("expected parent ID %v, got %v", parentID, *retrieved.ParentID)
	}
	assertCategoryAuditState(t, retrieved, userID, nil, nil)
}

// TestSaveAndDeleteMultiple tests saving and deleting multiple categories.
func TestSaveAndDeleteMultiple(t *testing.T) {
	ctx := context.Background()
	repo, err := catrepo.NewPostgresCategoryRepository(ctx, shared2.NewPostgresConfig().ConnectionString())
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}
	defer repo.Close()

	userID := shared.ActorID(uuid.New())

	// Create and save multiple categories
	categories := []*catalog.Category{
		catalog.NewCategory("Category 1", "Description 1"),
		catalog.NewCategory("Category 2", "Description 2"),
		catalog.NewCategory("Category 3", "Description 3"),
	}

	for _, cat := range categories {
		if err := repo.Save(ctx, *cat, userID); err != nil {
			t.Fatalf("failed to save category: %v", err)
		}
	}

	// Retrieve each one and verify
	for _, cat := range categories {
		retrieved, err := repo.Get(ctx, cat.ID)
		if err != nil {
			t.Fatalf("failed to retrieve category %v: %v", cat.ID, err)
		}

		if retrieved.Name != cat.Name {
			t.Errorf("expected name %q, got %q", cat.Name, retrieved.Name)
		}
		assertCategoryAuditState(t, retrieved, userID, nil, nil)
	}

	// Delete one
	if err := repo.Delete(ctx, categories[0].ID, userID); err != nil {
		t.Fatalf("failed to delete category: %v", err)
	}

	row := loadCategoryAuditRow(t, categories[0].ID)
	if !row.deletedAt.Valid {
		t.Fatal("expected deleted at to be set")
	}
	if !row.deletedBy.Valid {
		t.Fatal("expected deleted by to be set")
	}
	if row.deletedBy.UUID != uuid.UUID(userID) {
		t.Fatalf("expected deleted by %v, got %v", userID, row.deletedBy.UUID)
	}

	// Verify it's deleted
	_, err = repo.Get(ctx, categories[0].ID)
	if err == nil {
		t.Fatal("expected error for deleted category")
	}

	// Verify others still exist
	for i := 1; i < len(categories); i++ {
		_, err := repo.Get(ctx, categories[i].ID)
		if err != nil {
			t.Fatalf("expected category %d to exist after deleting another, got error: %v", i, err)
		}
	}
}

// TestSaveWithNonExistentParent tests that saving a category with a non-existent parent fails.
func TestSaveWithNonExistentParent(t *testing.T) {
	ctx := context.Background()
	repo, err := catrepo.NewPostgresCategoryRepository(ctx, shared2.NewPostgresConfig().ConnectionString())
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}
	defer repo.Close()

	userID := shared.ActorID(uuid.New())

	// Create a category with a non-existent parent
	cat := catalog.NewCategory("Child", "Orphan category")
	invalidParentID := catalog.ParentCategoryID(uuid.New())
	cat.MoveToParent(invalidParentID)

	// Try to save - should fail due to foreign key constraint
	err = repo.Save(ctx, *cat, userID)
	if err == nil {
		t.Fatal("expected error for non-existent parent, got nil")
	}
}

// TestDeleteNonExistentCategory tests that deleting a non-existent category fails gracefully.
func TestDeleteNonExistentCategory(t *testing.T) {
	ctx := context.Background()
	repo, err := catrepo.NewPostgresCategoryRepository(ctx, shared2.NewPostgresConfig().ConnectionString())
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}
	defer repo.Close()

	userID := shared.ActorID(uuid.New())
	nonExistentID := catalog.CategoryID(uuid.New())

	// Delete a category that doesn't exist - should not fail (idempotent)
	err = repo.Delete(ctx, nonExistentID, userID)
	if err != nil {
		t.Fatalf("expected delete to be idempotent, got error: %v", err)
	}
}
