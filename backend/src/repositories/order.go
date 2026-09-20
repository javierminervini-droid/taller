package repositories

import (
	"context"
	"strings"
	"time"

	"taller-gestion/backend/src/models"
	"taller-gestion/backend/src/utils"
)

const orderSelect = `
  SELECT o.id, o.client_id, o.technician_id, o.product_type_id, o.product_id, o.product_label,
         o.locality, o.provider_id, o.status_id, o.title, o.description,
         to_char(o.scheduled_date, 'YYYY-MM-DD') AS scheduled_date,
         o.scheduled_time, o.hours, o.km, o.parts_cost, o.parts_sale,
         o.started_at, o.completed_at, o.created_at, o.updated_at,
         c.name AS client_name, c.phone AS client_phone, c.email AS client_email,
         c.address AS client_address, c.created_at AS client_created_at, c.source AS client_source,
         t.name AS technician_name,
         pt.name AS product_type_name,
         p.name AS product_name,
         pr.name AS provider_name, pr.kind AS provider_kind,
         s.name AS status_name, s.color AS status_color,
         to_char(COALESCE(c.created_at, o.created_at)::date, 'YYYY-MM-DD') AS load_date,
         to_char(o.created_at::date, 'YYYY-MM-DD') AS entry_date,
         to_char(o.created_at, 'YYYY') AS entry_year,
         to_char(o.created_at, 'MM') AS entry_month
  FROM service_orders o
  JOIN clients c ON c.id = o.client_id
  JOIN service_statuses s ON s.id = o.status_id
  LEFT JOIN technicians t ON t.id = o.technician_id
  LEFT JOIN product_types pt ON pt.id = o.product_type_id
  LEFT JOIN products p ON p.id = o.product_id
  LEFT JOIN providers pr ON pr.id = o.provider_id
`

type OrderFilters struct {
	Date          string
	Year          string
	Month         string
	TechnicianID  string
	ProductTypeID string
	ProviderID    string
	Locality      string
	StatusID      string
	ClientID      string
}

func (q *OrderFilters) Where() (string, []any) {
	var clauses []string
	var params []any
	if q.Date != "" {
		clauses = append(clauses, "o.scheduled_date = ?")
		params = append(params, q.Date)
	}
	if q.Year != "" {
		clauses = append(clauses, "to_char(o.created_at, 'YYYY') = ?")
		params = append(params, q.Year)
	}
	if q.Month != "" {
		m := q.Month
		if len(m) == 1 {
			m = "0" + m
		}
		clauses = append(clauses, "to_char(o.created_at, 'MM') = ?")
		params = append(params, m)
	}
	if q.TechnicianID != "" {
		clauses = append(clauses, "o.technician_id = ?")
		params = append(params, utils.Atoi64(q.TechnicianID))
	}
	if q.ProductTypeID != "" {
		clauses = append(clauses, "o.product_type_id = ?")
		params = append(params, utils.Atoi64(q.ProductTypeID))
	}
	if q.ProviderID != "" {
		clauses = append(clauses, "o.provider_id = ?")
		params = append(params, utils.Atoi64(q.ProviderID))
	}
	if q.Locality != "" {
		clauses = append(clauses, "o.locality ILIKE ?")
		params = append(params, "%"+q.Locality+"%")
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
	sql := orderSelect + " " + where + " " + orderBy
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

func (r *Repos) FindOrderByID(ctx context.Context, id int64) (*models.ServiceOrder, error) {
	var order models.ServiceOrder
	err := r.DB.NewSelect().Model(&order).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *Repos) InsertOrder(ctx context.Context, order *models.ServiceOrder) error {
	_, err := r.DB.NewInsert().Model(order).Exec(ctx)
	return err
}

func (r *Repos) UpdateOrder(ctx context.Context, id int64, fields map[string]any, startedAt, completedAt *time.Time) error {
	_, err := r.DB.NewUpdate().Model((*models.ServiceOrder)(nil)).
		Set("client_id = ?", utils.MustInt64(fields["client_id"])).
		Set("technician_id = ?", utils.OptInt64(fields["technician_id"])).
		Set("product_type_id = ?", utils.OptInt64(fields["product_type_id"])).
		Set("product_id = ?", utils.OptInt64(fields["product_id"])).
		Set("product_label = ?", utils.OptStr(fields["product_label"])).
		Set("locality = ?", utils.OptStr(fields["locality"])).
		Set("provider_id = ?", utils.OptInt64(fields["provider_id"])).
		Set("status_id = ?", utils.MustInt64(fields["status_id"])).
		Set("title = ?", utils.CoalesceTitle(fields["title"])).
		Set("description = ?", utils.OptStr(fields["description"])).
		Set("scheduled_date = ?", utils.OptStr(fields["scheduled_date"])).
		Set("scheduled_time = ?", utils.OptStr(fields["scheduled_time"])).
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
		SELECT DISTINCT to_char(created_at, 'YYYY') AS year
		FROM service_orders
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
