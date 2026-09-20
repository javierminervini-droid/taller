package models

import "github.com/uptrace/bun"

type WhatsAppTemplate struct {
	bun.BaseModel `bun:"table:whatsapp_templates"`

	ID       int64  `bun:"id,pk,autoincrement" json:"id"`
	Name     string `bun:"name,notnull" json:"name"`
	Body     string `bun:"body,notnull" json:"body"`
	StatusID *int64 `bun:"status_id" json:"status_id"`
	AutoOpen bool   `bun:"auto_open,notnull" json:"auto_open"`
	Active   bool   `bun:"active,notnull" json:"active"`
}
