package repositories

import (
	"context"

	"taller-gestion/backend/src/models"
	"taller-gestion/backend/src/utils"
)

func (r *Repos) ListTariffs(ctx context.Context) ([]map[string]any, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT t.*, p.name AS provider_name, tech.name AS technician_name
		FROM tariffs t
		LEFT JOIN providers p ON p.id = t.provider_id
		LEFT JOIN technicians tech ON tech.id = t.technician_id
		ORDER BY t.scope, t.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return utils.RowsToMaps(rows)
}

func (r *Repos) ListTariffModels(ctx context.Context) ([]models.Tariff, error) {
	var list []models.Tariff
	err := r.DB.NewSelect().Model(&list).OrderExpr("scope, id").Scan(ctx)
	return list, err
}

func (r *Repos) InsertTariff(ctx context.Context, t *models.Tariff) error {
	_, err := r.DB.NewInsert().Model(t).Exec(ctx)
	return err
}

func (r *Repos) ListOrdersForResults(ctx context.Context, from, to string) ([]map[string]any, error) {
	var fromArg, toArg any
	if from != "" {
		fromArg = from
	}
	if to != "" {
		toArg = to
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT o.*, c.name AS client_name, t.name AS technician_name,
		       pr.name AS provider_name, pr.kind AS provider_kind,
		       to_char(o.visit_date, 'YYYY-MM-DD') AS scheduled_date,
		       to_char(o.visit_date, 'YYYY-MM-DD') AS visit_date
		FROM service_requests o
		JOIN clients c ON c.id = o.client_id
		LEFT JOIN technicians t ON t.id = o.technician_id
		LEFT JOIN providers pr ON pr.id = o.provider_id
		WHERE (? IS NULL OR o.visit_date >= ?::date)
		  AND (? IS NULL OR o.visit_date <= ?::date)
		ORDER BY o.visit_date, o.id`, fromArg, fromArg, toArg, toArg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return utils.RowsToMaps(rows)
}
