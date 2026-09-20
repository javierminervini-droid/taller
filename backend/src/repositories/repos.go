package repositories

import (
	"github.com/uptrace/bun"
)

type Repos struct {
	DB *bun.DB
}

func New(db *bun.DB) *Repos {
	return &Repos{DB: db}
}
