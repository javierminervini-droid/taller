package services

import (
	"context"

	"taller-gestion/backend/src/models"
	"taller-gestion/backend/src/utils"
)

func (s *Services) ListProducts(ctx context.Context) ([]map[string]any, error) {
	return s.Repos.ListProductsJoined(ctx)
}

func (s *Services) CreateProduct(ctx context.Context, claims *utils.Claims, b map[string]any) (int64, error) {
	if !utils.HasRole(claims.Role, "admin", "coordinador") {
		return 0, ErrForbidden
	}
	p := &models.Product{
		Name:          utils.StrOr(b["name"], ""),
		SKU:           utils.OptStr(b["sku"]),
		ProductTypeID: utils.OptInt64(b["product_type_id"]),
		ProviderID:    utils.OptInt64(b["provider_id"]),
		Stock:         int(utils.FloatOr(b["stock"], 0)),
		Price:         utils.FloatOr(b["price"], 0),
	}
	if err := s.Repos.InsertProduct(ctx, p); err != nil {
		return 0, err
	}
	return p.ID, nil
}
