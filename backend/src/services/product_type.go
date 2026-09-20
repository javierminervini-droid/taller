package services

import (
	"context"

	"taller-gestion/backend/src/models"
	"taller-gestion/backend/src/utils"
)

func (s *Services) CreateProductType(ctx context.Context, claims *utils.Claims, name string) (int64, error) {
	if !utils.HasRole(claims.Role, "admin", "coordinador") {
		return 0, ErrForbidden
	}
	pt := &models.ProductType{Name: name}
	if err := s.Repos.InsertProductType(ctx, pt); err != nil {
		return 0, err
	}
	return pt.ID, nil
}

func (s *Services) PatchProductType(ctx context.Context, claims *utils.Claims, id int64, name string) error {
	if !utils.HasRole(claims.Role, "admin", "coordinador") {
		return ErrForbidden
	}
	return s.Repos.UpdateProductTypeName(ctx, id, name)
}
