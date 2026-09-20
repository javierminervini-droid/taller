package controllers

import (
	"net/http"

	"taller-gestion/backend/src/middleware"
	"taller-gestion/backend/src/utils"

	"github.com/gin-gonic/gin"
)

func (d *Deps) ListWhatsAppTemplates(c *gin.Context) {
	claims := middleware.Claims(c)
	list, err := d.Services.ListWhatsAppTemplates(c.Request.Context(), claims)
	if writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, list)
}

func (d *Deps) CreateWhatsAppTemplate(c *gin.Context) {
	claims := middleware.Claims(c)
	var b map[string]any
	if err := c.ShouldBindJSON(&b); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	id, err := d.Services.CreateWhatsAppTemplate(c.Request.Context(), claims, b)
	if writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (d *Deps) PatchWhatsAppTemplate(c *gin.Context) {
	claims := middleware.Claims(c)
	id := utils.Atoi64(c.Param("id"))
	var b map[string]any
	_ = c.ShouldBindJSON(&b)
	if err := d.Services.PatchWhatsAppTemplate(c.Request.Context(), claims, id, b); writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (d *Deps) OrderWhatsApp(c *gin.Context) {
	claims := middleware.Claims(c)
	id := utils.Atoi64(c.Param("id"))
	var tplID *int64
	if q := c.Query("templateId"); q != "" {
		n := utils.Atoi64(q)
		tplID = &n
	}
	msg, err := d.Services.OrderWhatsApp(c.Request.Context(), claims, id, tplID)
	if writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, msg)
}
