package repositories

import (
	"context"

	"taller-gestion/backend/src/models"
)

func (r *Repos) ListStatuses(ctx context.Context) ([]models.ServiceStatus, error) {
	var list []models.ServiceStatus
	err := r.DB.NewSelect().Model(&list).OrderExpr("sort_order").Scan(ctx)
	return list, err
}

func (r *Repos) FindStatusByID(ctx context.Context, id int64) (*models.ServiceStatus, error) {
	var st models.ServiceStatus
	err := r.DB.NewSelect().Model(&st).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &st, nil
}
