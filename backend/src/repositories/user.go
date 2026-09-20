package repositories

import (
	"context"

	"taller-gestion/backend/src/models"
)

type UserPublic struct {
	ID       int64  `bun:"id" json:"id"`
	Username string `bun:"username" json:"username"`
	FullName string `bun:"full_name" json:"full_name"`
	Role     string `bun:"role" json:"role"`
	Active   bool   `bun:"active" json:"active"`
}

func (r *Repos) FindUserByUsernameActive(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.DB.NewSelect().Model(&user).
		Where("username = ? AND active = TRUE", username).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repos) FindUserByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	err := r.DB.NewSelect().Model(&user).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repos) ListUsersPublic(ctx context.Context, orderBy string) ([]UserPublic, error) {
	if orderBy == "" {
		orderBy = "full_name"
	}
	var users []UserPublic
	err := r.DB.NewSelect().
		TableExpr("users").
		Column("id", "username", "full_name", "role", "active").
		OrderExpr(orderBy).
		Scan(ctx, &users)
	if err != nil {
		return nil, err
	}
	if users == nil {
		users = []UserPublic{}
	}
	return users, nil
}

func (r *Repos) InsertUser(ctx context.Context, user *models.User) error {
	_, err := r.DB.NewInsert().Model(user).Exec(ctx)
	return err
}

func (r *Repos) FindFirstActiveUserByRole(ctx context.Context, role string) (*models.User, error) {
	var user models.User
	err := r.DB.NewSelect().Model(&user).
		Where("role = ? AND active = TRUE", role).
		OrderExpr("id").
		Limit(1).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
