package models

import "github.com/uptrace/bun"

type UnitType struct {
	bun.BaseModel `bun:"table:unit_types"`

	ID   int64  `bun:"id,pk,autoincrement" json:"id"`
	Name string `bun:"name,notnull" json:"name"`
}
