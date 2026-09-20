package controllers

import (
	"net/http"

	"taller-gestion/backend/src/middleware"

	"github.com/gin-gonic/gin"
)

func (d *Deps) Login(c *gin.Context) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	_ = c.ShouldBindJSON(&body)
	token, user, err := d.Services.Login(c.Request.Context(), body.Username, body.Password)
	if writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

func (d *Deps) Me(c *gin.Context) {
	claims := middleware.Claims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sesión requerida"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": d.Services.Me(c.Request.Context(), claims)})
}
