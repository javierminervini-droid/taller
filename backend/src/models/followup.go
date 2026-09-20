package models

import (
	"time"

	"github.com/uptrace/bun"
)

type Followup struct {
	bun.BaseModel `bun:"table:followups"`

	ID             int64      `bun:"id,pk,autoincrement" json:"id"`
	ServiceRequestID int64      `bun:"service_request_id,notnull" json:"service_request_id"`
	RuleID         *int64     `bun:"rule_id" json:"rule_id"`
	AssignedUserID *int64     `bun:"assigned_user_id" json:"assigned_user_id"`
	DueAt          *time.Time `bun:"due_at" json:"due_at"`
	Notes          *string    `bun:"notes" json:"notes"`
	Status         string     `bun:"status,notnull" json:"status"`
	CreatedAt      time.Time  `bun:"created_at,nullzero,notnull,default:now()" json:"created_at"`
	CompletedAt    *time.Time `bun:"completed_at" json:"completed_at"`
}
