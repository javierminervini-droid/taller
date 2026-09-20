package controllers

import (
	"net/http"

	"taller-gestion/backend/src/middleware"

	"github.com/gin-gonic/gin"
)

func (d *Deps) ListTariffs(c *gin.Context) {
	claims := middleware.Claims(c)
	list, err := d.Services.ListTariffs(c.Request.Context(), claims)
	if writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, list)
}

func (d *Deps) CreateTariff(c *gin.Context) {
	claims := middleware.Claims(c)
	var b map[string]any
	if err := c.ShouldBindJSON(&b); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	id, err := d.Services.CreateTariff(c.Request.Context(), claims, b)
	if writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (d *Deps) Results(c *gin.Context) {
	claims := middleware.Claims(c)
	result, err := d.Services.ResultsReport(
		c.Request.Context(), claims,
		c.Query("from"), c.Query("to"), c.Query("group"),
	)
	if writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, result)
}
