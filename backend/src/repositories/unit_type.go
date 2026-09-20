package repositories

import (
	"context"

	"taller-gestion/backend/src/models"
)

func (r *Repos) ListUnitTypes(ctx context.Context) ([]models.UnitType, error) {
	var list []models.UnitType
	err := r.DB.NewSelect().Model(&list).OrderExpr("name").Scan(ctx)
	return list, err
}

func (r *Repos) InsertUnitType(ctx context.Context, ut *models.UnitType) error {
	_, err := r.DB.NewInsert().Model(ut).Exec(ctx)
	return err
}

func (r *Repos) UpdateUnitTypeName(ctx context.Context, id int64, name string) error {
	_, err := r.DB.NewUpdate().Model((*models.UnitType)(nil)).
		Set("name = ?", name).
		Where("id = ?", id).
		Exec(ctx)
	return err
}
