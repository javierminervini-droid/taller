package controllers

import (
	"net/http"

	"taller-gestion/backend/src/middleware"

	"github.com/gin-gonic/gin"
)

func (d *Deps) Lookups(c *gin.Context) {
	claims := middleware.Claims(c)
	result, err := d.Services.Lookups(c.Request.Context(), claims)
	if writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, result)
}
