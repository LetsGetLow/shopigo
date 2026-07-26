package shared

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestActorIDFromNullUUID(t *testing.T) {
	t.Run("returns nil for invalid uuid", func(t *testing.T) {
		if got := ActorIDFromNullUUID(uuid.NullUUID{}); got != nil {
			t.Fatalf("expected nil, got %v", got)
		}
	})

	t.Run("returns actor id for valid uuid", func(t *testing.T) {
		value := uuid.New()
		got := ActorIDFromNullUUID(uuid.NullUUID{UUID: value, Valid: true})
		if got == nil {
			t.Fatal("expected non-nil actor id")
		}
		if uuid.UUID(*got) != value {
			t.Fatalf("expected %v, got %v", value, *got)
		}
	})
}

func TestUpdatedAtFromNullTime(t *testing.T) {
	t.Run("returns nil for invalid time", func(t *testing.T) {
		if got := UpdatedAtFromNullTime(sql.NullTime{}); got != nil {
			t.Fatalf("expected nil, got %v", got)
		}
	})

	t.Run("returns updated at for valid time", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Second)
		got := UpdatedAtFromNullTime(sql.NullTime{Time: now, Valid: true})
		if got == nil {
			t.Fatal("expected non-nil updated at")
		}
		if time.Time(*got) != now {
			t.Fatalf("expected %v, got %v", now, time.Time(*got))
		}
	})
}

func TestDeletedAtFromNullTime(t *testing.T) {
	t.Run("returns nil for invalid time", func(t *testing.T) {
		if got := DeletedAtFromNullTime(sql.NullTime{}); got != nil {
			t.Fatalf("expected nil, got %v", got)
		}
	})

	t.Run("returns deleted at for valid time", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Second)
		got := DeletedAtFromNullTime(sql.NullTime{Time: now, Valid: true})
		if got == nil {
			t.Fatal("expected non-nil deleted at")
		}
		if time.Time(*got) != now {
			t.Fatalf("expected %v, got %v", now, time.Time(*got))
		}
	})
}

func TestAuditAccessors(t *testing.T) {
	createdAt := CreatedAt(time.Date(2026, 7, 26, 10, 0, 0, 0, time.UTC))
	createdBy := ActorID(uuid.New())
	updatedAt := UpdatedAt(time.Date(2026, 7, 26, 11, 0, 0, 0, time.UTC))
	updatedBy := ActorID(uuid.New())
	deletedAt := DeletedAt(time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC))
	deletedBy := ActorID(uuid.New())

	audit := NewAudit(createdAt, createdBy, &updatedAt, &updatedBy, &deletedAt, &deletedBy)

	if audit.CreatedAt() != createdAt {
		t.Fatalf("expected createdAt %v, got %v", createdAt, audit.CreatedAt())
	}
	if audit.CreatedBy() != createdBy {
		t.Fatalf("expected createdBy %v, got %v", createdBy, audit.CreatedBy())
	}
	if audit.UpdatedAt() == nil || *audit.UpdatedAt() != updatedAt {
		t.Fatalf("expected updatedAt %v, got %v", updatedAt, audit.UpdatedAt())
	}
	if audit.UpdatedBy() == nil || *audit.UpdatedBy() != updatedBy {
		t.Fatalf("expected updatedBy %v, got %v", updatedBy, audit.UpdatedBy())
	}
	if audit.DeletedAt() == nil || *audit.DeletedAt() != deletedAt {
		t.Fatalf("expected deletedAt %v, got %v", deletedAt, audit.DeletedAt())
	}
	if audit.DeletedBy() == nil || *audit.DeletedBy() != deletedBy {
		t.Fatalf("expected deletedBy %v, got %v", deletedBy, audit.DeletedBy())
	}
	if !audit.IsDeleted() {
		t.Fatal("expected audit to be deleted")
	}
}

func TestAuditIsDeletedFalseWhenUnset(t *testing.T) {
	audit := NewAudit(CreatedAt(time.Now()), ActorID(uuid.New()), nil, nil, nil, nil)

	if audit.IsDeleted() {
		t.Fatal("expected audit to be not deleted")
	}
}
