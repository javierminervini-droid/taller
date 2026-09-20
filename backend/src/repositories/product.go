package repositories

import (
	"context"

	"taller-gestion/backend/src/models"
	"taller-gestion/backend/src/utils"
)

func (r *Repos) ListProducts(ctx context.Context) ([]models.Product, error) {
	var list []models.Product
	err := r.DB.NewSelect().Model(&list).OrderExpr("name").Scan(ctx)
	return list, err
}

func (r *Repos) ListProductsJoined(ctx context.Context) ([]map[string]any, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT p.*, pt.name AS type_name, pr.name AS provider_name
		FROM products p
		LEFT JOIN unit_types pt ON pt.id = p.product_type_id
		LEFT JOIN providers pr ON pr.id = p.provider_id
		ORDER BY p.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return utils.RowsToMaps(rows)
}

func (r *Repos) InsertProduct(ctx context.Context, p *models.Product) error {
	_, err := r.DB.NewInsert().Model(p).Exec(ctx)
	return err
}
