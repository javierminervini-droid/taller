package services

import (
	"context"

	"taller-gestion/backend/src/models"
	"taller-gestion/backend/src/utils"
)

func (s *Services) Login(ctx context.Context, username, password string) (token string, user utils.PublicUser, err error) {
	u, err := s.Repos.FindUserByUsernameActive(ctx, username)
	if err != nil || !utils.CheckPassword(u.PasswordHash, password) {
		return "", utils.PublicUser{}, ErrInvalidCredentials
	}
	token, err = utils.SignUser(s.JWTSecret, u)
	if err != nil {
		return "", utils.PublicUser{}, ErrSignToken
	}
	return token, utils.PublicUserFrom(ctx, s.Repos.DB, u), nil
}

func (s *Services) Me(ctx context.Context, claims *utils.Claims) utils.PublicUser {
	user, err := s.Repos.FindUserByID(ctx, claims.ID)
	if err != nil {
		return utils.PublicUserFromClaims(ctx, s.Repos.DB, claims)
	}
	return utils.PublicUserFrom(ctx, s.Repos.DB, user)
}

func (s *Services) TechnicianOf(ctx context.Context, claims *utils.Claims) *models.Technician {
	if claims == nil {
		return nil
	}
	return utils.TechnicianOf(ctx, s.Repos.DB, claims.ID, claims.Role)
}
