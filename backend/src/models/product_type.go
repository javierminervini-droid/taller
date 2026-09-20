package models

import "github.com/uptrace/bun"

type ProductType struct {
	bun.BaseModel `bun:"table:product_types"`

	ID   int64  `bun:"id,pk,autoincrement" json:"id"`
	Name string `bun:"name,notnull" json:"name"`
}
