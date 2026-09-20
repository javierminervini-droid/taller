package services

import (
	"context"

	"taller-gestion/backend/src/models"
	"taller-gestion/backend/src/repositories"
	"taller-gestion/backend/src/utils"
)

type LookupsResult struct {
	Technicians []models.Technician       `json:"technicians"`
	Providers   []models.Provider         `json:"providers"`
	UnitTypes   []models.UnitType         `json:"unitTypes"`
	ProductTypes []models.UnitType        `json:"productTypes"` // alias for older clients
	Statuses    []models.ServiceStatus    `json:"statuses"`
	Products    []models.Product          `json:"products"`
	Users       []repositories.UserPublic `json:"users"`
}

func (s *Services) Lookups(ctx context.Context, claims *utils.Claims) (*LookupsResult, error) {
	statuses, _ := s.Repos.ListStatuses(ctx)
	unitTypes, _ := s.Repos.ListUnitTypes(ctx)
	products, _ := s.Repos.ListProducts(ctx)

	if statuses == nil {
		statuses = []models.ServiceStatus{}
	}
	if unitTypes == nil {
		unitTypes = []models.UnitType{}
	}
	if products == nil {
		products = []models.Product{}
	}

	if claims != nil && claims.Role == "tecnico" {
		tech := s.TechnicianOf(ctx, claims)
		technicians := []models.Technician{}
		if tech != nil {
			technicians = append(technicians, *tech)
		}
		return &LookupsResult{
			Technicians:  technicians,
			Providers:    []models.Provider{},
			UnitTypes:    unitTypes,
			ProductTypes: unitTypes,
			Statuses:     statuses,
			Products:     products,
			Users:        []repositories.UserPublic{},
		}, nil
	}

	technicians, _ := s.Repos.ListActiveTechnicians(ctx)
	providers, _ := s.Repos.ListProviders(ctx)
	if providers == nil {
		providers = []models.Provider{}
	}
	users, _ := s.Repos.ListUsersPublic(ctx, "full_name")

	return &LookupsResult{
		Technicians:  technicians,
		Providers:    providers,
		UnitTypes:    unitTypes,
		ProductTypes: unitTypes,
		Statuses:     statuses,
		Products:     products,
		Users:        users,
	}, nil
}
