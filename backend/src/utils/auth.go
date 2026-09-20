package utils

import (
	"context"
	"errors"
	"time"

	"taller-gestion/backend/src/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	FullName string `json:"full_name"`
	jwt.RegisteredClaims
}

type PublicUser struct {
	ID             int64   `json:"id"`
	Username       string  `json:"username"`
	FullName       string  `json:"full_name"`
	Role           string  `json:"role"`
	TechnicianID   *int64  `json:"technician_id"`
	TechnicianName *string `json:"technician_name"`
}

func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(b), err
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func SignUser(secret string, user *models.User) (string, error) {
	claims := Claims{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
		FullName: user.FullName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(12 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

func VerifyToken(secret, tokenStr string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func TechnicianOf(ctx context.Context, db *bun.DB, userID int64, role string) *models.Technician {
	if role != "tecnico" {
		return nil
	}
	var tech models.Technician
	err := db.NewSelect().Model(&tech).
		Where("user_id = ? AND active = TRUE", userID).
		Scan(ctx)
	if err != nil {
		return nil
	}
	return &tech
}

func TechnicianOfAny(ctx context.Context, db *bun.DB, userID int64, role string) *models.Technician {
	if role != "tecnico" {
		return nil
	}
	var tech models.Technician
	err := db.NewSelect().Model(&tech).
		Where("user_id = ?", userID).
		Scan(ctx)
	if err != nil {
		return nil
	}
	return &tech
}

func PublicUserFrom(ctx context.Context, db *bun.DB, user *models.User) PublicUser {
	tech := TechnicianOf(ctx, db, user.ID, user.Role)
	if tech == nil {
		tech = TechnicianOfAny(ctx, db, user.ID, user.Role)
	}
	pu := PublicUser{
		ID:       user.ID,
		Username: user.Username,
		FullName: user.FullName,
		Role:     user.Role,
	}
	if tech != nil {
		id := tech.ID
		name := tech.Name
		pu.TechnicianID = &id
		pu.TechnicianName = &name
	}
	return pu
}

func PublicUserFromClaims(ctx context.Context, db *bun.DB, c *Claims) PublicUser {
	u := &models.User{
		ID:       c.ID,
		Username: c.Username,
		FullName: c.FullName,
		Role:     c.Role,
	}
	return PublicUserFrom(ctx, db, u)
}

func CanManage(role string) bool {
	return role == "admin" || role == "coordinador"
}

func HasRole(role string, roles ...string) bool {
	for _, r := range roles {
		if role == r {
			return true
		}
	}
	return false
}
