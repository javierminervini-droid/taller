package services

import (
	"context"

	"taller-gestion/backend/src/models"
	"taller-gestion/backend/src/repositories"
	"taller-gestion/backend/src/utils"
)

type LookupsResult struct {
	Technicians  []models.Technician       `json:"technicians"`
	Providers    []models.Provider         `json:"providers"`
	ProductTypes []models.ProductType      `json:"productTypes"`
	Statuses     []models.ServiceStatus    `json:"statuses"`
	Products     []models.Product          `json:"products"`
	Users        []repositories.UserPublic `json:"users"`
}

func (s *Services) Lookups(ctx context.Context, claims *utils.Claims) (*LookupsResult, error) {
	statuses, _ := s.Repos.ListStatuses(ctx)
	productTypes, _ := s.Repos.ListProductTypes(ctx)
	products, _ := s.Repos.ListProducts(ctx)

	if statuses == nil {
		statuses = []models.ServiceStatus{}
	}
	if productTypes == nil {
		productTypes = []models.ProductType{}
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
			ProductTypes: productTypes,
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
		ProductTypes: productTypes,
		Statuses:     statuses,
		Products:     products,
		Users:        users,
	}, nil
}
