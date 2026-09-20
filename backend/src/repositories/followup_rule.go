package repositories

import (
	"context"

	"taller-gestion/backend/src/models"
	"taller-gestion/backend/src/utils"
)

func (r *Repos) ListActiveFollowupRules(ctx context.Context) ([]models.FollowupRule, error) {
	var rules []models.FollowupRule
	err := r.DB.NewSelect().Model(&rules).Where("active = TRUE").Scan(ctx)
	return rules, err
}

func (r *Repos) ListFollowupRulesJoined(ctx context.Context) ([]map[string]any, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT r.*, s.name AS status_name FROM followup_rules r
		LEFT JOIN service_statuses s ON s.id = r.status_id
		ORDER BY r.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return utils.RowsToMaps(rows)
}

func (r *Repos) InsertFollowupRule(ctx context.Context, rule *models.FollowupRule) error {
	_, err := r.DB.NewInsert().Model(rule).Exec(ctx)
	return err
}
