package services

import (
	"context"

	"taller-gestion/backend/src/models"
	"taller-gestion/backend/src/repositories"
	"taller-gestion/backend/src/utils"
)

func (s *Services) ListUsers(ctx context.Context, claims *utils.Claims) ([]repositories.UserPublic, error) {
	if !utils.HasRole(claims.Role, "admin") {
		return nil, ErrForbidden
	}
	return s.Repos.ListUsersPublic(ctx, "id")
}

func (s *Services) CreateUser(ctx context.Context, claims *utils.Claims, username, password, fullName, role string, phone, specialty *string) (int64, error) {
	if !utils.HasRole(claims.Role, "admin") {
		return 0, ErrForbidden
	}
	hash, err := utils.HashPassword(password)
	if err != nil {
		return 0, err
	}
	user := &models.User{
		Username:     username,
		PasswordHash: hash,
		FullName:     fullName,
		Role:         role,
		Active:       true,
	}
	if err := s.Repos.InsertUser(ctx, user); err != nil {
		return 0, err
	}
	if role == "tecnico" {
		tech := &models.Technician{
			UserID:    &user.ID,
			Name:      fullName,
			Phone:     phone,
			Specialty: specialty,
			Active:    true,
		}
		if err := s.Repos.InsertTechnician(ctx, tech); err != nil {
			return 0, err
		}
	}
	return user.ID, nil
}
