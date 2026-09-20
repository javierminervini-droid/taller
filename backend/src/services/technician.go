package services

import (
	"context"

	"taller-gestion/backend/src/models"
	"taller-gestion/backend/src/utils"
)

func (s *Services) ListTechnicians(ctx context.Context) ([]models.Technician, error) {
	return s.Repos.ListTechnicians(ctx)
}

func (s *Services) CreateTechnician(ctx context.Context, claims *utils.Claims, b map[string]any) (int64, error) {
	if !utils.HasRole(claims.Role, "admin") {
		return 0, ErrForbidden
	}
	t := &models.Technician{
		UserID:    utils.OptInt64(b["user_id"]),
		Name:      utils.StrOr(b["name"], ""),
		Phone:     utils.OptStr(b["phone"]),
		Specialty: utils.OptStr(b["specialty"]),
		Active:    true,
	}
	if err := s.Repos.InsertTechnician(ctx, t); err != nil {
		return 0, err
	}
	return t.ID, nil
}
