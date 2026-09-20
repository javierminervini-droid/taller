package repositories

import (
	"context"

	"taller-gestion/backend/src/models"
)

func (r *Repos) ListProviders(ctx context.Context) ([]models.Provider, error) {
	var list []models.Provider
	err := r.DB.NewSelect().Model(&list).OrderExpr("name").Scan(ctx)
	return list, err
}

func (r *Repos) InsertProvider(ctx context.Context, p *models.Provider) error {
	_, err := r.DB.NewInsert().Model(p).Exec(ctx)
	return err
}
