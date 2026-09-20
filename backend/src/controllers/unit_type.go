package controllers

import (
	"net/http"

	"taller-gestion/backend/src/middleware"
	"taller-gestion/backend/src/utils"

	"github.com/gin-gonic/gin"
)

func (d *Deps) CreateUnitType(c *gin.Context) {
	claims := middleware.Claims(c)
	var b struct {
		Name string `json:"name"`
	}
	_ = c.ShouldBindJSON(&b)
	id, err := d.Services.CreateUnitType(c.Request.Context(), claims, b.Name)
	if writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (d *Deps) PatchUnitType(c *gin.Context) {
	claims := middleware.Claims(c)
	var b struct {
		Name string `json:"name"`
	}
	_ = c.ShouldBindJSON(&b)
	id := utils.Atoi64(c.Param("id"))
	if err := d.Services.PatchUnitType(c.Request.Context(), claims, id, b.Name); writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Legacy aliases
func (d *Deps) CreateProductType(c *gin.Context) { d.CreateUnitType(c) }
func (d *Deps) PatchProductType(c *gin.Context)  { d.PatchUnitType(c) }
