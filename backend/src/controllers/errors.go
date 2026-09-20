package controllers

import (
	"errors"
	"net/http"

	"taller-gestion/backend/src/services"

	"github.com/gin-gonic/gin"
)

func writeServiceError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, services.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "Sin permiso para esta acción"})
	case errors.Is(err, services.ErrFollowupsAdmin):
		c.JSON(http.StatusForbidden, gin.H{"error": "Los seguimientos los gestiona administración"})
	case errors.Is(err, services.ErrTechNotLinked):
		c.JSON(http.StatusForbidden, gin.H{"error": "Tu usuario no está vinculado a un perfil de técnico"})
	case errors.Is(err, services.ErrTechViewOrders):
		c.JSON(http.StatusForbidden, gin.H{"error": "Solo podés ver tus órdenes"})
	case errors.Is(err, services.ErrTechUpdateOrders):
		c.JSON(http.StatusForbidden, gin.H{"error": "Solo podés actualizar tus órdenes"})
	case errors.Is(err, services.ErrTechContactOrders):
		c.JSON(http.StatusForbidden, gin.H{"error": "Solo podés contactar clientes de tus órdenes"})
	case errors.Is(err, services.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Orden no encontrada"})
	case errors.Is(err, services.ErrClientNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Cliente no encontrado"})
	case errors.Is(err, services.ErrTemplateNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Plantilla no encontrada"})
	case errors.Is(err, services.ErrBadRequest):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario o clave incorrectos"})
	case errors.Is(err, services.ErrSignToken):
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo firmar el token"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	return true
}
