package domain

import "time"

type CreatedAt time.Time
type UpdatedAt time.Time
type DeletedAt time.Time

type Audit struct {
	createdAt *CreatedAt
	updatedAt *UpdatedAt
	deletedAt *DeletedAt
}

func NewAudit(
	ca *CreatedAt,
	ua *UpdatedAt,
	da *DeletedAt,
) Audit {
	return Audit{
		createdAt: ca,
		updatedAt: ua,
		deletedAt: da,
	}
}

func (a Audit) CreatedAt() *time.Time {
	return (*time.Time)(a.createdAt)
}

func (a Audit) UpdatedAt() *time.Time {
	return (*time.Time)(a.updatedAt)
}

func (a Audit) DeletedAt() *time.Time {
	return (*time.Time)(a.deletedAt)
}
