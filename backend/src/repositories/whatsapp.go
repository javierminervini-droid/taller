package repositories

import (
	"context"

	"taller-gestion/backend/src/models"
	"taller-gestion/backend/src/utils"
)

func (r *Repos) ListWhatsAppTemplates(ctx context.Context) ([]map[string]any, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT t.*, s.name AS status_name
		FROM whatsapp_templates t
		LEFT JOIN service_statuses s ON s.id = t.status_id
		ORDER BY t.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return utils.RowsToMaps(rows)
}

func (r *Repos) FindWhatsAppTemplateByID(ctx context.Context, id int64) (*models.WhatsAppTemplate, error) {
	var t models.WhatsAppTemplate
	err := r.DB.NewSelect().Model(&t).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repos) InsertWhatsAppTemplate(ctx context.Context, t *models.WhatsAppTemplate) error {
	_, err := r.DB.NewInsert().Model(t).Exec(ctx)
	return err
}

func (r *Repos) UpdateWhatsAppTemplate(ctx context.Context, id int64, name, body string, statusID *int64, autoOpen, active bool) error {
	_, err := r.DB.NewUpdate().Model((*models.WhatsAppTemplate)(nil)).
		Set("name = ?", name).
		Set("body = ?", body).
		Set("status_id = ?", statusID).
		Set("auto_open = ?", autoOpen).
		Set("active = ?", active).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *Repos) FindActiveTemplateByID(ctx context.Context, id int64) (*models.WhatsAppTemplate, error) {
	var t models.WhatsAppTemplate
	err := r.DB.NewSelect().Model(&t).Where("id = ? AND active = TRUE", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repos) FindActiveTemplateByStatus(ctx context.Context, statusID int64) (*models.WhatsAppTemplate, error) {
	var t models.WhatsAppTemplate
	err := r.DB.NewSelect().Model(&t).
		Where("active = TRUE AND status_id = ?", statusID).
		OrderExpr("id ASC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repos) FindActiveGenericTemplate(ctx context.Context) (*models.WhatsAppTemplate, error) {
	var t models.WhatsAppTemplate
	err := r.DB.NewSelect().Model(&t).
		Where("active = TRUE AND status_id IS NULL").
		OrderExpr("id ASC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
