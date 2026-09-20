package repositories

import (
	"context"
	"time"

	"taller-gestion/backend/src/models"
	"taller-gestion/backend/src/utils"
)

func (r *Repos) PendingFollowupExists(ctx context.Context, orderID, ruleID int64) (bool, error) {
	return r.DB.NewSelect().Model((*models.Followup)(nil)).
		Where("service_order_id = ? AND rule_id = ? AND status = 'pendiente'", orderID, ruleID).
		Exists(ctx)
}

func (r *Repos) InsertFollowup(ctx context.Context, f *models.Followup) error {
	_, err := r.DB.NewInsert().Model(f).Exec(ctx)
	return err
}

func (r *Repos) ListFollowups(ctx context.Context, status string) ([]map[string]any, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT f.*, o.title AS order_title,
		       to_char(o.scheduled_date, 'YYYY-MM-DD') AS scheduled_date,
		       c.name AS client_name,
		       u.full_name AS assigned_name, r.name AS rule_name, s.name AS order_status
		FROM followups f
		JOIN service_orders o ON o.id = f.service_order_id
		JOIN clients c ON c.id = o.client_id
		JOIN service_statuses s ON s.id = o.status_id
		LEFT JOIN users u ON u.id = f.assigned_user_id
		LEFT JOIN followup_rules r ON r.id = f.rule_id
		WHERE (? = 'todos' OR f.status = ?)
		ORDER BY f.status ASC, f.created_at DESC`, status, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return utils.RowsToMaps(rows)
}

func (r *Repos) UpdateFollowupStatus(ctx context.Context, id int64, status string, completedAt *time.Time) error {
	if completedAt != nil {
		_, err := r.DB.NewUpdate().Model((*models.Followup)(nil)).
			Set("status = ?", status).
			Set("completed_at = ?", *completedAt).
			Where("id = ?", id).
			Exec(ctx)
		return err
	}
	_, err := r.DB.NewUpdate().Model((*models.Followup)(nil)).
		Set("status = ?", status).
		Set("completed_at = NULL").
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *Repos) OrderIDsForStatusTrigger(ctx context.Context, statusID int64) ([]int64, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT o.id FROM service_orders o
		JOIN service_statuses s ON s.id = o.status_id
		WHERE o.status_id = ? AND s.is_closed = FALSE`, statusID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *Repos) OrderIDsForElapsedTrigger(ctx context.Context, statusID int64, hours int) ([]int64, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT o.id FROM service_orders o
		JOIN service_statuses s ON s.id = o.status_id
		WHERE o.status_id = ? AND s.is_closed = FALSE
		  AND COALESCE(o.started_at, o.created_at) <= NOW() - make_interval(hours => ?)`,
		statusID, hours)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
