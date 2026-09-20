package models

import "github.com/uptrace/bun"

type Product struct {
	bun.BaseModel `bun:"table:products"`

	ID            int64   `bun:"id,pk,autoincrement" json:"id"`
	Name          string  `bun:"name,notnull" json:"name"`
	SKU           *string `bun:"sku" json:"sku"`
	ProductTypeID *int64  `bun:"product_type_id" json:"product_type_id"`
	ProviderID    *int64  `bun:"provider_id" json:"provider_id"`
	Stock         int     `bun:"stock,notnull" json:"stock"`
	Price         float64 `bun:"price,notnull" json:"price"`
}
