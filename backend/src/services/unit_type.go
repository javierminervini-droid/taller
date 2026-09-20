package services

import (
	"context"

	"taller-gestion/backend/src/models"
	"taller-gestion/backend/src/utils"
)

func (s *Services) CreateUnitType(ctx context.Context, claims *utils.Claims, name string) (int64, error) {
	if !utils.HasRole(claims.Role, "admin", "coordinador") {
		return 0, ErrForbidden
	}
	ut := &models.UnitType{Name: name}
	if err := s.Repos.InsertUnitType(ctx, ut); err != nil {
		return 0, err
	}
	return ut.ID, nil
}

func (s *Services) PatchUnitType(ctx context.Context, claims *utils.Claims, id int64, name string) error {
	if !utils.HasRole(claims.Role, "admin", "coordinador") {
		return ErrForbidden
	}
	return s.Repos.UpdateUnitTypeName(ctx, id, name)
}
