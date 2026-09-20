package models

import "github.com/uptrace/bun"

type FollowupRule struct {
	bun.BaseModel `bun:"table:followup_rules"`

	ID          int64   `bun:"id,pk,autoincrement" json:"id"`
	Name        string  `bun:"name,notnull" json:"name"`
	TriggerType string  `bun:"trigger_type,notnull" json:"trigger_type"`
	StatusID    *int64  `bun:"status_id" json:"status_id"`
	Hours       *int    `bun:"hours" json:"hours"`
	AssignRole  *string `bun:"assign_role" json:"assign_role"`
	Active      bool    `bun:"active,notnull" json:"active"`
}
