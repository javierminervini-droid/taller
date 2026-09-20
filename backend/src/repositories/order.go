package repositories

import (
	"context"
	"strings"
	"time"

	"taller-gestion/backend/src/models"
	"taller-gestion/backend/src/utils"
)

const requestSelect = `
  SELECT o.id, o.client_id, o.technician_id, o.unit_type_id, o.product_id, o.product_label,
         o.locality, o.provider_id, o.status_id, o.title, o.description,
         to_char(o.received_at, 'YYYY-MM-DD') AS received_at,
         o.provider_order_ref, o.internal_order_no, o.request_kind,
         o.appliance_model, o.reported_failure, o.ops_notes, o.diagnosis_notes,
         to_char(o.visit_date, 'YYYY-MM-DD') AS visit_date,
         o.scheduled_time, o.hours, o.km, o.parts_cost, o.parts_sale,
         o.started_at, o.completed_at, o.created_at, o.updated_at,
         c.name AS client_name, c.phone AS client_phone, c.phone_alt AS client_phone_alt,
         c.email AS client_email,
         c.address AS client_address, c.locality AS client_locality,
         c.created_at AS client_created_at, c.source AS client_source,
         t.name AS technician_name, t.code AS technician_code,
         ut.name AS unit_type_name,
         p.name AS product_name,
         pr.name AS provider_name, pr.kind AS provider_kind,
         s.name AS status_name, s.color AS status_color,
         to_char(COALESCE(o.received_at, c.created_at::date, o.created_at::date), 'YYYY-MM-DD') AS load_date,
         to_char(COALESCE(o.received_at, o.created_at::date), 'YYYY-MM-DD') AS entry_date,
         to_char(COALESCE(o.received_at, o.created_at::date), 'YYYY') AS entry_year,
         to_char(COALESCE(o.received_at, o.created_at::date), 'MM') AS entry_month
  FROM service_requests o
  JOIN clients c ON c.id = o.client_id
  JOIN service_statuses s ON s.id = o.status_id
  LEFT JOIN technicians t ON t.id = o.technician_id
  LEFT JOIN unit_types ut ON ut.id = o.unit_type_id
  LEFT JOIN products p ON p.id = o.product_id
  LEFT JOIN providers pr ON pr.id = o.provider_id
`

type OrderFilters struct {
	Date         string
	Year         string
	Month        string
	TechnicianID string
	UnitTypeID   string
	ProviderID   string
	Locality     string
	StatusID     string
	ClientID     string
}

func (q *OrderFilters) Where() (string, []any) {
	var clauses []string
	var params []any
	if q.Date != "" {
		clauses = append(clauses, "o.visit_date = ?")
		params = append(params, q.Date)
	}
	if q.Year != "" {
		clauses = append(clauses, "to_char(COALESCE(o.received_at, o.created_at::date), 'YYYY') = ?")
		params = append(params, q.Year)
	}
	if q.Month != "" {
		m := q.Month
		if len(m) == 1 {
			m = "0" + m
		}
		clauses = append(clauses, "to_char(COALESCE(o.received_at, o.created_at::date), 'MM') = ?")
		params = append(params, m)
	}
	if q.TechnicianID != "" {
		clauses = append(clauses, "o.technician_id = ?")
		params = append(params, utils.Atoi64(q.TechnicianID))
	}
	if q.UnitTypeID != "" {
		clauses = append(clauses, "o.unit_type_id = ?")
		params = append(params, utils.Atoi64(q.UnitTypeID))
	}
	if q.ProviderID != "" {
		clauses = append(clauses, "o.provider_id = ?")
		params = append(params, utils.Atoi64(q.ProviderID))
	}
	if q.Locality != "" {
		clauses = append(clauses, "(o.locality ILIKE ? OR c.locality ILIKE ?)")
		params = append(params, "%"+q.Locality+"%", "%"+q.Locality+"%")
	}
	if q.StatusID != "" {
		clauses = append(clauses, "o.status_id = ?")
		params = append(params, utils.Atoi64(q.StatusID))
	}
	if q.ClientID != "" {
		clauses = append(clauses, "o.client_id = ?")
		params = append(params, utils.Atoi64(q.ClientID))
	}
	if len(clauses) == 0 {
		return "", nil
	}
	return "WHERE " + strings.Join(clauses, " AND "), params
}

func (r *Repos) QueryOrders(ctx context.Context, where string, params []any, orderBy string) ([]map[string]any, error) {
	sql := requestSelect + " " + where + " " + orderBy
	rows, err := r.DB.QueryContext(ctx, sql, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return utils.RowsToMaps(rows)
}

func (r *Repos) QueryOrderByID(ctx context.Context, id int64) (map[string]any, error) {
	list, err := r.QueryOrders(ctx, "WHERE o.id = ?", []any{id}, "")
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}
	return list[0], nil
}

func (r *Repos) FindOrderByID(ctx context.Context, id int64) (*models.ServiceRequest, error) {
	var order models.ServiceRequest
	err := r.DB.NewSelect().Model(&order).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *Repos) InsertOrder(ctx context.Context, order *models.ServiceRequest) error {
	_, err := r.DB.NewInsert().Model(order).Exec(ctx)
	return err
}

func (r *Repos) UpdateOrder(ctx context.Context, id int64, fields map[string]any, startedAt, completedAt *time.Time) error {
	_, err := r.DB.NewUpdate().Model((*models.ServiceRequest)(nil)).
		Set("client_id = ?", utils.MustInt64(fields["client_id"])).
		Set("technician_id = ?", utils.OptInt64(fields["technician_id"])).
		Set("unit_type_id = ?", utils.OptInt64(fields["unit_type_id"])).
		Set("product_id = ?", utils.OptInt64(fields["product_id"])).
		Set("product_label = ?", utils.OptStr(fields["product_label"])).
		Set("locality = ?", utils.OptStr(fields["locality"])).
		Set("provider_id = ?", utils.OptInt64(fields["provider_id"])).
		Set("status_id = ?", utils.MustInt64(fields["status_id"])).
		Set("title = ?", utils.CoalesceTitle(fields["title"])).
		Set("description = ?", utils.OptStr(fields["description"])).
		Set("received_at = ?", utils.OptStr(fields["received_at"])).
		Set("provider_order_ref = ?", utils.OptStr(fields["provider_order_ref"])).
		Set("internal_order_no = ?", utils.OptStr(fields["internal_order_no"])).
		Set("request_kind = ?", utils.OptStr(fields["request_kind"])).
		Set("appliance_model = ?", utils.OptStr(fields["appliance_model"])).
		Set("reported_failure = ?", utils.OptStr(fields["reported_failure"])).
		Set("visit_date = ?", utils.OptStr(fields["visit_date"])).
		Set("scheduled_time = ?", utils.OptStr(fields["scheduled_time"])).
		Set("ops_notes = ?", utils.OptStr(fields["ops_notes"])).
		Set("diagnosis_notes = ?", utils.OptStr(fields["diagnosis_notes"])).
		Set("hours = ?", utils.FloatOr(fields["hours"], 0)).
		Set("km = ?", utils.FloatOr(fields["km"], 0)).
		Set("parts_cost = ?", utils.FloatOr(fields["parts_cost"], 0)).
		Set("parts_sale = ?", utils.FloatOr(fields["parts_sale"], 0)).
		Set("started_at = ?", startedAt).
		Set("completed_at = ?", completedAt).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *Repos) ListOrderYears(ctx context.Context) ([]string, error) {
	years := []string{}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT DISTINCT to_char(COALESCE(received_at, created_at::date), 'YYYY') AS year
		FROM service_requests
		WHERE created_at IS NOT NULL
		ORDER BY year DESC`)
	if err != nil {
		return years, err
	}
	defer rows.Close()
	for rows.Next() {
		var y string
		if rows.Scan(&y) == nil && y != "" {
			years = append(years, y)
		}
	}
	return years, rows.Err()
}
