package controllers

import (
	"taller-gestion/backend/src/services"
)

type Deps struct {
	Services *services.Services
}

func New(svcs *services.Services) *Deps {
	return &Deps{Services: svcs}
}
