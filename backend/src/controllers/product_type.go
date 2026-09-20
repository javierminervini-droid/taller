package controllers

import (
	"net/http"

	"taller-gestion/backend/src/middleware"
	"taller-gestion/backend/src/utils"

	"github.com/gin-gonic/gin"
)

func (d *Deps) CreateProductType(c *gin.Context) {
	claims := middleware.Claims(c)
	var b struct {
		Name string `json:"name"`
	}
	_ = c.ShouldBindJSON(&b)
	id, err := d.Services.CreateProductType(c.Request.Context(), claims, b.Name)
	if writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (d *Deps) PatchProductType(c *gin.Context) {
	claims := middleware.Claims(c)
	var b struct {
		Name string `json:"name"`
	}
	_ = c.ShouldBindJSON(&b)
	id := utils.Atoi64(c.Param("id"))
	if err := d.Services.PatchProductType(c.Request.Context(), claims, id, b.Name); writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
