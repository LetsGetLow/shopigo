package shared

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type CreatedAt time.Time
type UpdatedAt time.Time
type DeletedAt time.Time

type ActorID uuid.UUID

func ActorIDFromNullUUID(id uuid.NullUUID) *ActorID {
	if !id.Valid {
		return nil
	}
	actorID := ActorID(id.UUID)
	return &actorID
}

func UpdatedAtFromNullTime(t sql.NullTime) *UpdatedAt {
	if !t.Valid {
		return nil
	}
	updatedAt := UpdatedAt(t.Time)
	return &updatedAt
}

func DeletedAtFromNullTime(t sql.NullTime) *DeletedAt {
	if !t.Valid {
		return nil
	}
	deletedAt := DeletedAt(t.Time)
	return &deletedAt
}

type Audit struct {
	createdAt CreatedAt
	createdBy ActorID
	updatedAt *UpdatedAt
	updatedBy *ActorID
	deletedAt *DeletedAt
	deletedBy *ActorID
}

func NewAudit(
	ca CreatedAt,
	cb ActorID,
	ua *UpdatedAt,
	ub *ActorID,
	da *DeletedAt,
	db *ActorID,
) Audit {
	return Audit{
		createdAt: ca,
		createdBy: cb,
		updatedAt: ua,
		updatedBy: ub,
		deletedAt: da,
		deletedBy: db,
	}
}

func (a Audit) CreatedAt() CreatedAt {
	return a.createdAt
}

func (a Audit) CreatedBy() ActorID {
	return a.createdBy
}

func (a Audit) UpdatedAt() *UpdatedAt {
	return a.updatedAt
}

func (a Audit) UpdatedBy() *ActorID {
	return a.updatedBy
}

func (a Audit) DeletedAt() *DeletedAt {
	return a.deletedAt
}

func (a Audit) DeletedBy() *ActorID {
	return a.deletedBy
}

func (s Audit) IsDeleted() bool {
	return s.deletedAt != nil
}
