package controllers

import (
	"net/http"

	"taller-gestion/backend/src/middleware"
	"taller-gestion/backend/src/repositories"
	"taller-gestion/backend/src/utils"

	"github.com/gin-gonic/gin"
)

func parseOrderFilters(c *gin.Context) repositories.OrderFilters {
	return repositories.OrderFilters{
		Date:          c.Query("date"),
		Year:          c.Query("year"),
		Month:         c.Query("month"),
		TechnicianID:  c.Query("technicianId"),
		ProductTypeID: c.Query("productTypeId"),
		ProviderID:    c.Query("providerId"),
		Locality:      c.Query("locality"),
		StatusID:      c.Query("statusId"),
		ClientID:      c.Query("clientId"),
	}
}

func (d *Deps) Agenda(c *gin.Context) {
	claims := middleware.Claims(c)
	result, err := d.Services.Agenda(c.Request.Context(), claims, parseOrderFilters(c))
	if writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, result)
}

func (d *Deps) ListOrders(c *gin.Context) {
	claims := middleware.Claims(c)
	result, err := d.Services.ListOrders(c.Request.Context(), claims, parseOrderFilters(c))
	if writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, result)
}

func (d *Deps) GetOrder(c *gin.Context) {
	claims := middleware.Claims(c)
	id := utils.Atoi64(c.Param("id"))
	order, err := d.Services.GetOrder(c.Request.Context(), claims, id)
	if writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, order)
}

func (d *Deps) CreateOrder(c *gin.Context) {
	claims := middleware.Claims(c)
	var b map[string]any
	if err := c.ShouldBindJSON(&b); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	id, err := d.Services.CreateOrder(c.Request.Context(), claims, b)
	if writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (d *Deps) PatchOrder(c *gin.Context) {
	claims := middleware.Claims(c)
	id := utils.Atoi64(c.Param("id"))
	var b map[string]any
	_ = c.ShouldBindJSON(&b)
	wa, err := d.Services.PatchOrder(c.Request.Context(), claims, id, b)
	if writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "whatsapp": wa})
}
