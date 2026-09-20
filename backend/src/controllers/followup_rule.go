package controllers

import (
	"net/http"

	"taller-gestion/backend/src/middleware"

	"github.com/gin-gonic/gin"
)

func (d *Deps) ListFollowupRules(c *gin.Context) {
	claims := middleware.Claims(c)
	list, err := d.Services.ListFollowupRules(c.Request.Context(), claims)
	if writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, list)
}

func (d *Deps) CreateFollowupRule(c *gin.Context) {
	claims := middleware.Claims(c)
	var b map[string]any
	if err := c.ShouldBindJSON(&b); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	id, err := d.Services.CreateFollowupRule(c.Request.Context(), claims, b)
	if writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}
