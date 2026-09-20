package controllers

import (
	"io"
	"net/http"

	"taller-gestion/backend/src/middleware"
	"taller-gestion/backend/src/utils"

	"github.com/gin-gonic/gin"
)

func (d *Deps) ListClients(c *gin.Context) {
	claims := middleware.Claims(c)
	list, err := d.Services.ListClients(c.Request.Context(), claims, c.Query("providerId"), c.Query("source"))
	if writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, list)
}

func (d *Deps) CreateClient(c *gin.Context) {
	claims := middleware.Claims(c)
	var b map[string]any
	if err := c.ShouldBindJSON(&b); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	id, err := d.Services.CreateClient(c.Request.Context(), claims, b)
	if writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (d *Deps) PatchClient(c *gin.Context) {
	claims := middleware.Claims(c)
	id := utils.Atoi64(c.Param("id"))
	var b map[string]any
	_ = c.ShouldBindJSON(&b)
	if err := d.Services.PatchClient(c.Request.Context(), claims, id, b); writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (d *Deps) ClientsTemplate(c *gin.Context) {
	claims := middleware.Claims(c)
	buf, err := d.Services.ClientsTemplate(c.Request.Context(), claims)
	if writeServiceError(c, err) {
		return
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="plantilla-clientes-salesforce.xlsx"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf)
}

func (d *Deps) ImportClients(c *gin.Context) {
	claims := middleware.Claims(c)
	providerID := utils.Atoi64(c.Query("providerId"))
	if providerID == 0 {
		providerID = utils.Atoi64(c.PostForm("provider_id"))
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Adjuntá el Excel de Salesforce"})
		return
	}
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No se pudo leer el Excel. Usá la plantilla del sistema."})
		return
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No se pudo leer el Excel. Usá la plantilla del sistema."})
		return
	}
	result, err := d.Services.ImportClients(c.Request.Context(), claims, providerID, data)
	if writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, result)
}
