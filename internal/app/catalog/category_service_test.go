package catalog_test

import (
	"context"
	"errors"
	"testing"

	appcatalog "shopigo/internal/app/catalog"
	domain "shopigo/internal/domain/catalog"
	domainshared "shopigo/internal/domain/shared"

	"github.com/google/uuid"
)

type fakeCategoryRepo struct {
	categories map[domain.CategoryID]*domain.Category

	saveErr   error
	getErr    error
	listErr   error
	listByErr error
	deleteErr error

	savedCategory   *domain.Category
	savedUser       domainshared.ActorID
	deletedID       domain.CategoryID
	deletedUser     domainshared.ActorID
	listByParentArg *domain.ParentCategoryID
}

func newFakeCategoryRepo() *fakeCategoryRepo {
	return &fakeCategoryRepo{
		categories: make(map[domain.CategoryID]*domain.Category),
	}
}

func (f *fakeCategoryRepo) Save(ctx context.Context, category domain.Category, user domainshared.ActorID) error {
	if f.saveErr != nil {
		return f.saveErr
	}

	cat := category
	f.categories[cat.ID] = &cat
	f.savedCategory = &cat
	f.savedUser = user
	return nil
}

func (f *fakeCategoryRepo) Get(ctx context.Context, id domain.CategoryID) (*domain.Category, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}

	category, ok := f.categories[id]
	if !ok {
		return nil, errors.New("category not found")
	}

	cat := *category
	return &cat, nil
}

func (f *fakeCategoryRepo) List(ctx context.Context) ([]domain.Category, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}

	out := make([]domain.Category, 0, len(f.categories))
	for _, category := range f.categories {
		out = append(out, *category)
	}
	return out, nil
}

func (f *fakeCategoryRepo) ListByParent(ctx context.Context, id *domain.ParentCategoryID) ([]domain.Category, error) {
	if f.listByErr != nil {
		return nil, f.listByErr
	}

	f.listByParentArg = id
	out := make([]domain.Category, 0)
	for _, category := range f.categories {
		switch {
		case id == nil && category.ParentID == nil:
			out = append(out, *category)
		case id != nil && category.ParentID != nil && *category.ParentID == *id:
			out = append(out, *category)
		}
	}
	return out, nil
}

func (f *fakeCategoryRepo) Delete(ctx context.Context, id domain.CategoryID, user domainshared.ActorID) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}

	f.deletedID = id
	f.deletedUser = user
	delete(f.categories, id)
	return nil
}

func (f *fakeCategoryRepo) Close() error { return nil }

func TestCategoryService_CreateCategory(t *testing.T) {
	repo := newFakeCategoryRepo()
	service := appcatalog.NewCategoryService(repo)
	userID := domainshared.ActorID(uuid.New())
	parentID := uuid.NewString()

	got, err := service.CreateCategory(context.Background(), appcatalog.CreateCategoryRequest{
		Name:        "Electronics",
		Description: "Gadgets and devices",
		ParentID:    &parentID,
	}, userID)
	if err != nil {
		t.Fatalf("expected create to succeed, got error: %v", err)
	}

	if got.Category.Name != "Electronics" || got.Category.Description != "Gadgets and devices" {
		t.Fatalf("unexpected category: %+v", got.Category)
	}
	if got.Category.ParentID == nil || *got.Category.ParentID != parentID {
		t.Fatalf("expected parent id %q, got %v", parentID, got.Category.ParentID)
	}
	if repo.savedCategory == nil {
		t.Fatal("expected repository save to be called")
	}
	if repo.savedUser != userID {
		t.Fatalf("expected save user %v, got %v", userID, repo.savedUser)
	}
}

func TestCategoryService_UpdateCategory(t *testing.T) {
	repo := newFakeCategoryRepo()
	existing := domain.NewCategory("Books", "Reading")
	repo.categories[existing.ID] = existing
	service := appcatalog.NewCategoryService(repo)
	userID := domainshared.ActorID(uuid.New())
	parentID := uuid.NewString()

	got, err := service.UpdateCategory(context.Background(), appcatalog.UpdateCategoryRequest{
		ID:          uuid.UUID(existing.ID).String(),
		Name:        "Books Updated",
		Description: "Updated reading",
		ParentID:    &parentID,
	}, userID)
	if err != nil {
		t.Fatalf("expected update to succeed, got error: %v", err)
	}
	if got.Category.Name != "Books Updated" {
		t.Fatalf("expected updated name, got %q", got.Category.Name)
	}
	if got.Category.Description != "Updated reading" {
		t.Fatalf("expected updated description, got %q", got.Category.Description)
	}
	if got.Category.ParentID == nil || *got.Category.ParentID != parentID {
		t.Fatalf("expected parent id %q, got %v", parentID, got.Category.ParentID)
	}
	if repo.savedCategory == nil || repo.savedCategory.Name != "Books Updated" {
		t.Fatal("expected updated category to be saved")
	}
}

func TestCategoryService_UpdateCategoryToRoot(t *testing.T) {
	repo := newFakeCategoryRepo()
	existing := domain.NewCategory("Books", "Reading")
	parentID := domain.ParentCategoryID(uuid.New())
	existing.MoveToParent(parentID)
	repo.categories[existing.ID] = existing
	service := appcatalog.NewCategoryService(repo)
	userID := domainshared.ActorID(uuid.New())
	emptyParent := ""

	got, err := service.UpdateCategory(context.Background(), appcatalog.UpdateCategoryRequest{
		ID:          uuid.UUID(existing.ID).String(),
		Name:        existing.Name,
		Description: existing.Description,
		ParentID:    &emptyParent,
	}, userID)
	if err != nil {
		t.Fatalf("expected update to root to succeed, got error: %v", err)
	}
	if got.Category.ParentID != nil {
		t.Fatalf("expected category to be root, got %v", got.Category.ParentID)
	}
}

func TestCategoryService_MoveCategoryToParent(t *testing.T) {
	repo := newFakeCategoryRepo()
	existing := domain.NewCategory("Laptops", "Portable computers")
	repo.categories[existing.ID] = existing
	service := appcatalog.NewCategoryService(repo)
	userID := domainshared.ActorID(uuid.New())
	parentID := domain.ParentCategoryID(uuid.New())

	got, err := service.MoveCategoryToParent(context.Background(), uuid.UUID(existing.ID).String(), uuid.UUID(parentID).String(), userID)
	if err != nil {
		t.Fatalf("expected move to succeed, got error: %v", err)
	}
	if got.ParentID == nil || *got.ParentID != uuid.UUID(parentID).String() {
		t.Fatalf("expected parent %v, got %v", parentID, got.ParentID)
	}
}

func TestCategoryService_MakeCategoryRoot(t *testing.T) {
	repo := newFakeCategoryRepo()
	existing := domain.NewCategory("Phones", "Mobile phones")
	parentID := domain.ParentCategoryID(uuid.New())
	existing.MoveToParent(parentID)
	repo.categories[existing.ID] = existing
	service := appcatalog.NewCategoryService(repo)
	userID := domainshared.ActorID(uuid.New())

	got, err := service.MakeCategoryRoot(context.Background(), uuid.UUID(existing.ID).String(), userID)
	if err != nil {
		t.Fatalf("expected make-root to succeed, got error: %v", err)
	}
	if got.ParentID != nil {
		t.Fatal("expected category to be root")
	}
}

func TestCategoryService_DeleteCategory(t *testing.T) {
	repo := newFakeCategoryRepo()
	existing := domain.NewCategory("Delete me", "Soon gone")
	repo.categories[existing.ID] = existing
	service := appcatalog.NewCategoryService(repo)
	userID := domainshared.ActorID(uuid.New())

	got, err := service.DeleteCategory(context.Background(), appcatalog.DeleteCategoryRequest{
		ID: uuid.UUID(existing.ID).String(),
	}, userID)
	if err != nil {
		t.Fatalf("expected delete to succeed, got error: %v", err)
	}
	if got.ID != uuid.UUID(existing.ID).String() {
		t.Fatalf("expected delete id %v, got %v", existing.ID, got.ID)
	}
	if repo.deletedID != existing.ID {
		t.Fatalf("expected delete id %v, got %v", existing.ID, repo.deletedID)
	}
}

func TestCategoryService_PropagatesRepositoryErrors(t *testing.T) {
	repo := newFakeCategoryRepo()
	repo.saveErr = errors.New("save failed")
	service := appcatalog.NewCategoryService(repo)

	_, err := service.CreateCategory(context.Background(), appcatalog.CreateCategoryRequest{
		Name:        "Broken",
		Description: "Category",
	}, domainshared.ActorID(uuid.New()))
	if err == nil || err.Error() != "save failed" {
		t.Fatalf("expected save error to propagate, got %v", err)
	}
}

func TestCategoryService_ListAndGetForwardToRepository(t *testing.T) {
	repo := newFakeCategoryRepo()
	category := domain.NewCategory("Root", "Root category")
	repo.categories[category.ID] = category
	service := appcatalog.NewCategoryService(repo)

	got, err := service.GetCategory(context.Background(), uuid.UUID(category.ID).String())
	if err != nil {
		t.Fatalf("expected get to succeed, got error: %v", err)
	}
	if got.ID != uuid.UUID(category.ID).String() {
		t.Fatalf("expected category %v, got %v", category.ID, got.ID)
	}

	listed, err := service.ListCategories(context.Background())
	if err != nil {
		t.Fatalf("expected list to succeed, got error: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected 1 category, got %d", len(listed))
	}
}
