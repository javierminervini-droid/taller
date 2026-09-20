package controllers

import (
	"net/http"

	"taller-gestion/backend/src/middleware"

	"github.com/gin-gonic/gin"
)

func (d *Deps) ListUsers(c *gin.Context) {
	claims := middleware.Claims(c)
	users, err := d.Services.ListUsers(c.Request.Context(), claims)
	if writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, users)
}

func (d *Deps) CreateUser(c *gin.Context) {
	claims := middleware.Claims(c)
	var b struct {
		Username  string  `json:"username"`
		Password  string  `json:"password"`
		FullName  string  `json:"full_name"`
		Role      string  `json:"role"`
		Phone     *string `json:"phone"`
		Specialty *string `json:"specialty"`
	}
	if err := c.ShouldBindJSON(&b); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	id, err := d.Services.CreateUser(c.Request.Context(), claims, b.Username, b.Password, b.FullName, b.Role, b.Phone, b.Specialty)
	if writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}
