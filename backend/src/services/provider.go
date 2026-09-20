package services

import (
	"context"

	"taller-gestion/backend/src/models"
	"taller-gestion/backend/src/utils"
)

func (s *Services) ListProviders(ctx context.Context) ([]models.Provider, error) {
	list, err := s.Repos.ListProviders(ctx)
	if err != nil {
		return nil, err
	}
	if list == nil {
		return []models.Provider{}, nil
	}
	return list, nil
}

func (s *Services) CreateProvider(ctx context.Context, claims *utils.Claims, b map[string]any) (int64, error) {
	if !utils.HasRole(claims.Role, "admin") {
		return 0, ErrForbidden
	}
	p := &models.Provider{
		Name:    utils.StrOr(b["name"], ""),
		Kind:    utils.StrOr(b["kind"], ""),
		Contact: utils.OptStr(b["contact"]),
		Notes:   utils.OptStr(b["notes"]),
	}
	if err := s.Repos.InsertProvider(ctx, p); err != nil {
		return 0, err
	}
	return p.ID, nil
}
