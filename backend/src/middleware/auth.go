package middleware

import (
	"net/http"
	"strings"

	"taller-gestion/backend/src/utils"

	"github.com/gin-gonic/gin"
)

const UserKey = "user"

func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
		path := c.Request.URL.Path
		if path == "/api/auth/login" || path == "/health" {
			c.Next()
			return
		}
		if !strings.HasPrefix(path, "/api/") {
			c.Next()
			return
		}

		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Sesión requerida"})
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := utils.VerifyToken(jwtSecret, token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Sesión inválida"})
			return
		}
		c.Set(UserKey, claims)
		c.Next()
	}
}

func Claims(c *gin.Context) *utils.Claims {
	v, ok := c.Get(UserKey)
	if !ok {
		return nil
	}
	claims, _ := v.(*utils.Claims)
	return claims
}

func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := Claims(c)
		if claims == nil || !utils.HasRole(claims.Role, roles...) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Sin permiso para esta acción"})
			return
		}
		c.Next()
	}
}
