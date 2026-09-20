package models

import (
	"time"

	"github.com/uptrace/bun"
)

type Client struct {
	bun.BaseModel `bun:"table:clients"`

	ID         int64     `bun:"id,pk,autoincrement" json:"id"`
	ProviderID *int64    `bun:"provider_id" json:"provider_id"`
	ExternalID *string   `bun:"external_id" json:"external_id"`
	Name       string    `bun:"name,notnull" json:"name"`
	Phone      *string   `bun:"phone" json:"phone"`
	PhoneAlt   *string   `bun:"phone_alt" json:"phone_alt"`
	Email      *string   `bun:"email" json:"email"`
	Locality   *string   `bun:"locality" json:"locality"`
	Address    *string   `bun:"address" json:"address"`
	Notes      *string   `bun:"notes" json:"notes"`
	Source     string    `bun:"source,notnull" json:"source"`
	CreatedAt  time.Time `bun:"created_at,nullzero,notnull,default:now()" json:"created_at"`
}
