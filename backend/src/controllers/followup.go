package controllers

import (
	"net/http"

	"taller-gestion/backend/src/middleware"
	"taller-gestion/backend/src/utils"

	"github.com/gin-gonic/gin"
)

func (d *Deps) ListFollowups(c *gin.Context) {
	claims := middleware.Claims(c)
	list, err := d.Services.ListFollowups(c.Request.Context(), claims, c.Query("status"))
	if writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, list)
}

func (d *Deps) PatchFollowup(c *gin.Context) {
	id := utils.Atoi64(c.Param("id"))
	var b struct {
		Status string `json:"status"`
	}
	_ = c.ShouldBindJSON(&b)
	if err := d.Services.PatchFollowup(c.Request.Context(), id, b.Status); writeServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
