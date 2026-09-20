package models

import (
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users"`

	ID           int64     `bun:"id,pk,autoincrement" json:"id"`
	Username     string    `bun:"username,notnull" json:"username"`
	PasswordHash string    `bun:"password_hash,notnull" json:"-"`
	FullName     string    `bun:"full_name,notnull" json:"full_name"`
	Role         string    `bun:"role,notnull" json:"role"`
	Active       bool      `bun:"active,notnull" json:"active"`
	CreatedAt    time.Time `bun:"created_at,nullzero,notnull,default:now()" json:"created_at"`
}
