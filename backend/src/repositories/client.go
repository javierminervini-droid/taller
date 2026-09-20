package repositories

import (
	"context"

	"taller-gestion/backend/src/models"
	"taller-gestion/backend/src/utils"
)

func (r *Repos) ListClientsForTechnician(ctx context.Context, techID int64) ([]map[string]any, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT DISTINCT c.*, p.name AS provider_name, p.kind AS provider_kind
		FROM clients c
		JOIN service_requests o ON o.client_id = c.id
		LEFT JOIN providers p ON p.id = c.provider_id
		WHERE o.technician_id = ?
		ORDER BY c.name`, techID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return utils.RowsToMaps(rows)
}

func (r *Repos) ListClientsFiltered(ctx context.Context, providerID, source string) ([]map[string]any, error) {
	var clauses []string
	var params []any
	if providerID == "own" {
		clauses = append(clauses, "c.provider_id IS NULL")
	} else if providerID != "" {
		clauses = append(clauses, "c.provider_id = ?")
		params = append(params, utils.Atoi64(providerID))
	}
	if source != "" {
		clauses = append(clauses, "c.source = ?")
		params = append(params, source)
	}
	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + joinAND(clauses)
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT c.*, p.name AS provider_name, p.kind AS provider_kind
		FROM clients c
		LEFT JOIN providers p ON p.id = c.provider_id
		`+where+`
		ORDER BY c.name`, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return utils.RowsToMaps(rows)
}

func (r *Repos) FindClientByID(ctx context.Context, id int64) (*models.Client, error) {
	var client models.Client
	err := r.DB.NewSelect().Model(&client).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (r *Repos) InsertClient(ctx context.Context, client *models.Client) error {
	_, err := r.DB.NewInsert().Model(client).Exec(ctx)
	return err
}

func (r *Repos) UpdateClient(ctx context.Context, id int64, providerID *int64, externalID *string, name string, phone, phoneAlt, email, locality *string, address, notes *string) error {
	_, err := r.DB.NewUpdate().Model((*models.Client)(nil)).
		Set("provider_id = ?", providerID).
		Set("external_id = ?", externalID).
		Set("name = ?", name).
		Set("phone = ?", phone).
		Set("phone_alt = ?", phoneAlt).
		Set("email = ?", email).
		Set("locality = ?", locality).
		Set("address = ?", address).
		Set("notes = ?", notes).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *Repos) FindClientByProviderExternal(ctx context.Context, providerID int64, externalID string) (*models.Client, error) {
	var client models.Client
	err := r.DB.NewSelect().Model(&client).
		Where("provider_id = ? AND external_id = ?", providerID, externalID).
		Limit(1).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (r *Repos) UpdateClientImport(ctx context.Context, id int64, name string, phone, email, locality, address, notes *string) error {
	_, err := r.DB.NewUpdate().Model((*models.Client)(nil)).
		Set("name = ?", name).
		Set("phone = ?", phone).
		Set("email = ?", email).
		Set("locality = ?", locality).
		Set("address = ?", address).
		Set("notes = ?", notes).
		Set("source = ?", "excel").
		Where("id = ?", id).
		Exec(ctx)
	return err
}
