package services

import (
	"taller-gestion/backend/src/repositories"
)

type Services struct {
	Repos    *repositories.Repos
	JWTSecret string
}

func New(repos *repositories.Repos, jwtSecret string) *Services {
	return &Services{Repos: repos, JWTSecret: jwtSecret}
}
