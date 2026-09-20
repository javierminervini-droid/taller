package services

import (
	"context"
	"fmt"
	"math"
	"sort"

	"taller-gestion/backend/src/models"
	"taller-gestion/backend/src/utils"
)

func (s *Services) ListTariffs(ctx context.Context, claims *utils.Claims) ([]map[string]any, error) {
	if !utils.HasRole(claims.Role, "admin", "coordinador") {
		return nil, ErrForbidden
	}
	list, err := s.Repos.ListTariffs(ctx)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []map[string]any{}
	}
	return list, nil
}

func (s *Services) CreateTariff(ctx context.Context, claims *utils.Claims, b map[string]any) (int64, error) {
	if !utils.HasRole(claims.Role, "admin", "coordinador") {
		return 0, ErrForbidden
	}
	scope := utils.StrOr(b["scope"], "")
	if scope != "proveedor" && scope != "tecnico" && scope != "general" {
		return 0, BadRequest("scope inválido")
	}
	t := &models.Tariff{
		Name:           utils.StrOr(b["name"], ""),
		Scope:          scope,
		ProviderID:     utils.OptInt64(b["provider_id"]),
		TechnicianID:   utils.OptInt64(b["technician_id"]),
		IncomeFixed:    utils.FloatOr(b["income_fixed"], 0),
		IncomePerHour:  utils.FloatOr(b["income_per_hour"], 0),
		IncomePerKm:    utils.FloatOr(b["income_per_km"], 0),
		IncomePartsPct: utils.FloatOr(b["income_parts_pct"], 0),
		CostFixed:      utils.FloatOr(b["cost_fixed"], 0),
		CostPerHour:    utils.FloatOr(b["cost_per_hour"], 0),
		CostPerKm:      utils.FloatOr(b["cost_per_km"], 0),
		CostPartsPct:   utils.FloatOr(b["cost_parts_pct"], 100),
		Active:         true,
	}
	if t.Name == "" {
		return 0, BadRequest("nombre requerido")
	}
	if err := s.Repos.InsertTariff(ctx, t); err != nil {
		return 0, err
	}
	return t.ID, nil
}

type ResultsReport struct {
	Group  string           `json:"group"`
	From   any              `json:"from"`
	To     any              `json:"to"`
	Rows   []ResultsBucket  `json:"rows"`
	Totals ResultsTotals    `json:"totals"`
	Orders []map[string]any `json:"orders"`
}

type ResultsBucket struct {
	Key     string  `json:"key"`
	Label   string  `json:"label"`
	Orders  int     `json:"orders"`
	Income  float64 `json:"income"`
	Cost    float64 `json:"cost"`
	Profit  float64 `json:"profit"`
}

type ResultsTotals struct {
	Orders int     `json:"orders"`
	Income float64 `json:"income"`
	Cost   float64 `json:"cost"`
	Profit float64 `json:"profit"`
}

func money(n float64) float64 {
	return math.Round(n*100) / 100
}

func pickIncome(order map[string]any, tariffs []models.Tariff) models.Tariff {
	pid, _ := utils.AsInt64(order["provider_id"])
	for _, t := range tariffs {
		if t.Active && t.Scope == "proveedor" && t.ProviderID != nil && *t.ProviderID == pid {
			return t
		}
	}
	for _, t := range tariffs {
		if t.Active && t.Scope == "general" {
			return t
		}
	}
	return models.Tariff{}
}

func pickCost(order map[string]any, tariffs []models.Tariff) models.Tariff {
	tid, hasTech := utils.AsInt64(order["technician_id"])
	for _, t := range tariffs {
		if t.Active && t.Scope == "tecnico" && t.TechnicianID != nil && hasTech && *t.TechnicianID == tid {
			return t
		}
	}
	for _, t := range tariffs {
		if t.Active && t.Scope == "tecnico" && t.TechnicianID == nil {
			return t
		}
	}
	for _, t := range tariffs {
		if t.Active && t.Scope == "general" {
			return t
		}
	}
	return models.Tariff{}
}

func quoteOrder(order map[string]any, tariffs []models.Tariff) map[string]any {
	hours := utils.FloatOr(order["hours"], 0)
	km := utils.FloatOr(order["km"], 0)
	partsCost := utils.FloatOr(order["parts_cost"], 0)
	partsSale := utils.FloatOr(order["parts_sale"], 0)
	incomeT := pickIncome(order, tariffs)
	costT := pickCost(order, tariffs)
	income := money(
		incomeT.IncomeFixed +
			hours*incomeT.IncomePerHour +
			km*incomeT.IncomePerKm +
			partsSale*incomeT.IncomePartsPct/100,
	)
	cost := money(
		costT.CostFixed +
			hours*costT.CostPerHour +
			km*costT.CostPerKm +
			partsCost*costT.CostPartsPct/100,
	)
	out := make(map[string]any, len(order)+5)
	for k, v := range order {
		out[k] = v
	}
	out["income"] = income
	out["cost"] = cost
	out["profit"] = money(income - cost)
	if incomeT.Name != "" {
		out["income_tariff"] = incomeT.Name
	} else {
		out["income_tariff"] = nil
	}
	if costT.Name != "" {
		out["cost_tariff"] = costT.Name
	} else {
		out["cost_tariff"] = nil
	}
	return out
}

func (s *Services) ResultsReport(ctx context.Context, claims *utils.Claims, from, to, group string) (*ResultsReport, error) {
	if !utils.HasRole(claims.Role, "admin", "coordinador") {
		return nil, ErrForbidden
	}
	if group == "" {
		group = "proveedor"
	}
	tariffs, err := s.Repos.ListTariffModels(ctx)
	if err != nil {
		return nil, err
	}
	orders, err := s.Repos.ListOrdersForResults(ctx, from, to)
	if err != nil {
		return nil, err
	}
	quoted := make([]map[string]any, 0, len(orders))
	for _, o := range orders {
		quoted = append(quoted, quoteOrder(o, tariffs))
	}

	buckets := map[string]*ResultsBucket{}
	for _, o := range quoted {
		key := "sin-asignar"
		label := "Sin asignar"
		switch group {
		case "cliente":
			key = fmt.Sprint(o["client_id"])
			label = fmt.Sprint(o["client_name"])
		case "tecnico":
			if tid, ok := utils.AsInt64(o["technician_id"]); ok && tid > 0 {
				key = fmt.Sprint(tid)
				if n := fmt.Sprint(o["technician_name"]); n != "" && n != "<nil>" {
					label = n
				} else {
					label = "Sin técnico"
				}
			} else {
				key = "0"
				label = "Sin técnico"
			}
		default:
			if pid, ok := utils.AsInt64(o["provider_id"]); ok && pid > 0 {
				key = fmt.Sprint(pid)
				name := fmt.Sprint(o["provider_name"])
				kind := fmt.Sprint(o["provider_kind"])
				if name != "" && name != "<nil>" {
					label = name
					if kind != "" && kind != "<nil>" {
						label = name + " (" + kind + ")"
					}
				}
			} else {
				key = "0"
				label = "Carga propia / sin origen"
			}
		}
		b, ok := buckets[key]
		if !ok {
			b = &ResultsBucket{Key: key, Label: label}
			buckets[key] = b
		}
		b.Orders++
		b.Income = money(b.Income + utils.FloatOr(o["income"], 0))
		b.Cost = money(b.Cost + utils.FloatOr(o["cost"], 0))
		b.Profit = money(b.Profit + utils.FloatOr(o["profit"], 0))
	}

	rows := make([]ResultsBucket, 0, len(buckets))
	for _, b := range buckets {
		rows = append(rows, *b)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Profit > rows[j].Profit })

	totals := ResultsTotals{}
	for _, r := range rows {
		totals.Orders += r.Orders
		totals.Income = money(totals.Income + r.Income)
		totals.Cost = money(totals.Cost + r.Cost)
		totals.Profit = money(totals.Profit + r.Profit)
	}

	var fromOut, toOut any
	if from != "" {
		fromOut = from
	}
	if to != "" {
		toOut = to
	}
	return &ResultsReport{
		Group:  group,
		From:   fromOut,
		To:     toOut,
		Rows:   rows,
		Totals: totals,
		Orders: quoted,
	}, nil
}
