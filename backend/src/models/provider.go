package models

import "github.com/uptrace/bun"

type Provider struct {
	bun.BaseModel `bun:"table:providers"`

	ID      int64   `bun:"id,pk,autoincrement" json:"id"`
	Name    string  `bun:"name,notnull" json:"name"`
	Kind    string  `bun:"kind,notnull" json:"kind"`
	Contact *string `bun:"contact" json:"contact"`
	Notes   *string `bun:"notes" json:"notes"`
}
