package models

import "github.com/uptrace/bun"

type ServiceStatus struct {
	bun.BaseModel `bun:"table:service_statuses"`

	ID        int64  `bun:"id,pk,autoincrement" json:"id"`
	Name      string `bun:"name,notnull" json:"name"`
	SortOrder int    `bun:"sort_order,notnull" json:"sort_order"`
	Color     string `bun:"color,notnull" json:"color"`
	IsClosed  bool   `bun:"is_closed,notnull" json:"is_closed"`
}
