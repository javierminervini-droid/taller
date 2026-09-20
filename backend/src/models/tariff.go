package models

import "github.com/uptrace/bun"

type Tariff struct {
	bun.BaseModel `bun:"table:tariffs"`

	ID             int64   `bun:"id,pk,autoincrement" json:"id"`
	Name           string  `bun:"name,notnull" json:"name"`
	Scope          string  `bun:"scope,notnull" json:"scope"`
	ProviderID     *int64  `bun:"provider_id" json:"provider_id"`
	TechnicianID   *int64  `bun:"technician_id" json:"technician_id"`
	IncomeFixed    float64 `bun:"income_fixed,notnull" json:"income_fixed"`
	IncomePerHour  float64 `bun:"income_per_hour,notnull" json:"income_per_hour"`
	IncomePerKm    float64 `bun:"income_per_km,notnull" json:"income_per_km"`
	IncomePartsPct float64 `bun:"income_parts_pct,notnull" json:"income_parts_pct"`
	CostFixed      float64 `bun:"cost_fixed,notnull" json:"cost_fixed"`
	CostPerHour    float64 `bun:"cost_per_hour,notnull" json:"cost_per_hour"`
	CostPerKm      float64 `bun:"cost_per_km,notnull" json:"cost_per_km"`
	CostPartsPct   float64 `bun:"cost_parts_pct,notnull" json:"cost_parts_pct"`
	Active         bool    `bun:"active,notnull" json:"active"`
}
