package models

import "github.com/uptrace/bun"

type Technician struct {
	bun.BaseModel `bun:"table:technicians"`

	ID        int64   `bun:"id,pk,autoincrement" json:"id"`
	UserID    *int64  `bun:"user_id" json:"user_id"`
	Name      string  `bun:"name,notnull" json:"name"`
	Phone     *string `bun:"phone" json:"phone"`
	Specialty *string `bun:"specialty" json:"specialty"`
	Active    bool    `bun:"active,notnull" json:"active"`
}
