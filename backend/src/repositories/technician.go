package repositories

import (
	"context"

	"taller-gestion/backend/src/models"
)

func (r *Repos) FindActiveTechnicianByUserID(ctx context.Context, userID int64) (*models.Technician, error) {
	var tech models.Technician
	err := r.DB.NewSelect().Model(&tech).
		Where("user_id = ? AND active = TRUE", userID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &tech, nil
}

func (r *Repos) FindTechnicianByUserID(ctx context.Context, userID int64) (*models.Technician, error) {
	var tech models.Technician
	err := r.DB.NewSelect().Model(&tech).
		Where("user_id = ?", userID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &tech, nil
}

func (r *Repos) ListActiveTechnicians(ctx context.Context) ([]models.Technician, error) {
	var list []models.Technician
	err := r.DB.NewSelect().Model(&list).Where("active = TRUE").OrderExpr("name").Scan(ctx)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []models.Technician{}
	}
	return list, nil
}

func (r *Repos) ListTechnicians(ctx context.Context) ([]models.Technician, error) {
	var list []models.Technician
	err := r.DB.NewSelect().Model(&list).OrderExpr("name").Scan(ctx)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []models.Technician{}
	}
	return list, nil
}

func (r *Repos) InsertTechnician(ctx context.Context, t *models.Technician) error {
	_, err := r.DB.NewInsert().Model(t).Exec(ctx)
	return err
}
