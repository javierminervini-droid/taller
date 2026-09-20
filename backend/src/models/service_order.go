package models

import (
	"time"

	"github.com/uptrace/bun"
)

type ServiceOrder struct {
	bun.BaseModel `bun:"table:service_orders"`

	ID            int64      `bun:"id,pk,autoincrement" json:"id"`
	ClientID      int64      `bun:"client_id,notnull" json:"client_id"`
	TechnicianID  *int64     `bun:"technician_id" json:"technician_id"`
	ProductTypeID *int64     `bun:"product_type_id" json:"product_type_id"`
	ProductID     *int64     `bun:"product_id" json:"product_id"`
	ProductLabel  *string    `bun:"product_label" json:"product_label"`
	Locality      *string    `bun:"locality" json:"locality"`
	ProviderID    *int64     `bun:"provider_id" json:"provider_id"`
	StatusID      int64      `bun:"status_id,notnull" json:"status_id"`
	Title         string     `bun:"title,notnull" json:"title"`
	Description   *string    `bun:"description" json:"description"`
	ScheduledDate *string    `bun:"scheduled_date" json:"scheduled_date"`
	ScheduledTime *string    `bun:"scheduled_time" json:"scheduled_time"`
	Hours         float64    `bun:"hours,notnull" json:"hours"`
	Km            float64    `bun:"km,notnull" json:"km"`
	PartsCost     float64    `bun:"parts_cost,notnull" json:"parts_cost"`
	PartsSale     float64    `bun:"parts_sale,notnull" json:"parts_sale"`
	StartedAt     *time.Time `bun:"started_at" json:"started_at"`
	CompletedAt   *time.Time `bun:"completed_at" json:"completed_at"`
	CreatedAt     time.Time  `bun:"created_at,nullzero,notnull,default:now()" json:"created_at"`
	UpdatedAt     time.Time  `bun:"updated_at,nullzero,notnull,default:now()" json:"updated_at"`
}
