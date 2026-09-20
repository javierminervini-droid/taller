package repositories

import (
	"context"

	"taller-gestion/backend/src/models"
)

func (r *Repos) ListProductTypes(ctx context.Context) ([]models.ProductType, error) {
	var list []models.ProductType
	err := r.DB.NewSelect().Model(&list).OrderExpr("name").Scan(ctx)
	return list, err
}

func (r *Repos) InsertProductType(ctx context.Context, pt *models.ProductType) error {
	_, err := r.DB.NewInsert().Model(pt).Exec(ctx)
	return err
}

func (r *Repos) UpdateProductTypeName(ctx context.Context, id int64, name string) error {
	_, err := r.DB.NewUpdate().Model((*models.ProductType)(nil)).
		Set("name = ?", name).
		Where("id = ?", id).
		Exec(ctx)
	return err
}
