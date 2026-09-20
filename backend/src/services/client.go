package services

import (
	"context"

	"taller-gestion/backend/src/models"
	"taller-gestion/backend/src/utils"
)

func (s *Services) ListClients(ctx context.Context, claims *utils.Claims, providerID, source string) ([]map[string]any, error) {
	if claims.Role == "tecnico" {
		tech := s.TechnicianOf(ctx, claims)
		techID := int64(-1)
		if tech != nil {
			techID = tech.ID
		}
		return s.Repos.ListClientsForTechnician(ctx, techID)
	}
	return s.Repos.ListClientsFiltered(ctx, providerID, source)
}

func (s *Services) CreateClient(ctx context.Context, claims *utils.Claims, b map[string]any) (int64, error) {
	if !utils.HasRole(claims.Role, "admin", "coordinador") {
		return 0, ErrForbidden
	}
	client := &models.Client{
		ProviderID: utils.OptInt64(b["provider_id"]),
		ExternalID: utils.OptStr(b["external_id"]),
		Name:       utils.StrOr(b["name"], ""),
		Phone:      utils.OptStr(b["phone"]),
		PhoneAlt:   utils.OptStr(b["phone_alt"]),
		Email:      utils.OptStr(b["email"]),
		Locality:   utils.OptStr(b["locality"]),
		Address:    utils.OptStr(b["address"]),
		Notes:      utils.OptStr(b["notes"]),
		Source:     "manual",
	}
	if err := s.Repos.InsertClient(ctx, client); err != nil {
		return 0, err
	}
	return client.ID, nil
}

func (s *Services) PatchClient(ctx context.Context, claims *utils.Claims, id int64, b map[string]any) error {
	if !utils.HasRole(claims.Role, "admin", "coordinador") {
		return ErrForbidden
	}
	current, err := s.Repos.FindClientByID(ctx, id)
	if err != nil {
		return ErrClientNotFound
	}
	if b == nil {
		b = map[string]any{}
	}
	providerID := current.ProviderID
	if v, ok := b["provider_id"]; ok {
		providerID = utils.OptInt64(v)
	}
	externalID := current.ExternalID
	if v, ok := b["external_id"]; ok {
		externalID = utils.OptStr(v)
	}
	name := current.Name
	if v, ok := b["name"]; ok {
		name = utils.StrOr(v, name)
	}
	phone := current.Phone
	if v, ok := b["phone"]; ok {
		phone = utils.OptStr(v)
	}
	phoneAlt := current.PhoneAlt
	if v, ok := b["phone_alt"]; ok {
		phoneAlt = utils.OptStr(v)
	}
	email := current.Email
	if v, ok := b["email"]; ok {
		email = utils.OptStr(v)
	}
	locality := current.Locality
	if v, ok := b["locality"]; ok {
		locality = utils.OptStr(v)
	}
	address := current.Address
	if v, ok := b["address"]; ok {
		address = utils.OptStr(v)
	}
	notes := current.Notes
	if v, ok := b["notes"]; ok {
		notes = utils.OptStr(v)
	}
	return s.Repos.UpdateClient(ctx, id, providerID, externalID, name, phone, phoneAlt, email, locality, address, notes)
}
