package models

import (
	"time"

	"github.com/uptrace/bun"
)

type ServiceRequest struct {
	bun.BaseModel `bun:"table:service_requests"`

	ID               int64      `bun:"id,pk,autoincrement" json:"id"`
	ClientID         int64      `bun:"client_id,notnull" json:"client_id"`
	TechnicianID     *int64     `bun:"technician_id" json:"technician_id"`
	UnitTypeID       *int64     `bun:"unit_type_id" json:"unit_type_id"`
	ProductID        *int64     `bun:"product_id" json:"product_id"`
	ProductLabel     *string    `bun:"product_label" json:"product_label"`
	Locality         *string    `bun:"locality" json:"locality"`
	ProviderID       *int64     `bun:"provider_id" json:"provider_id"`
	StatusID         int64      `bun:"status_id,notnull" json:"status_id"`
	Title            string     `bun:"title,notnull" json:"title"`
	Description      *string    `bun:"description" json:"description"`
	ReceivedAt       *string    `bun:"received_at" json:"received_at"`
	ProviderOrderRef *string    `bun:"provider_order_ref" json:"provider_order_ref"`
	InternalOrderNo  *string    `bun:"internal_order_no" json:"internal_order_no"`
	RequestKind      *string    `bun:"request_kind" json:"request_kind"`
	ApplianceModel   *string    `bun:"appliance_model" json:"appliance_model"`
	ReportedFailure  *string    `bun:"reported_failure" json:"reported_failure"`
	VisitDate        *string    `bun:"visit_date" json:"visit_date"`
	ScheduledTime    *string    `bun:"scheduled_time" json:"scheduled_time"`
	OpsNotes         *string    `bun:"ops_notes" json:"ops_notes"`
	DiagnosisNotes   *string    `bun:"diagnosis_notes" json:"diagnosis_notes"`
	Hours            float64    `bun:"hours,notnull" json:"hours"`
	Km               float64    `bun:"km,notnull" json:"km"`
	PartsCost        float64    `bun:"parts_cost,notnull" json:"parts_cost"`
	PartsSale        float64    `bun:"parts_sale,notnull" json:"parts_sale"`
	StartedAt        *time.Time `bun:"started_at" json:"started_at"`
	CompletedAt      *time.Time `bun:"completed_at" json:"completed_at"`
	CreatedAt        time.Time  `bun:"created_at,nullzero,notnull,default:now()" json:"created_at"`
	UpdatedAt        time.Time  `bun:"updated_at,nullzero,notnull,default:now()" json:"updated_at"`
}
