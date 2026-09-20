package controllers

import (
	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine, deps *Deps) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	apiGroup := r.Group("/api")
	{
		apiGroup.POST("/auth/login", deps.Login)
		apiGroup.GET("/auth/me", deps.Me)

		apiGroup.GET("/lookups", deps.Lookups)

		apiGroup.GET("/agenda", deps.Agenda)
		apiGroup.GET("/service-requests", deps.ListOrders)
		apiGroup.POST("/service-requests", deps.CreateOrder)
		apiGroup.GET("/service-requests/:id/whatsapp", deps.OrderWhatsApp)
		apiGroup.GET("/service-requests/:id", deps.GetOrder)
		apiGroup.PATCH("/service-requests/:id", deps.PatchOrder)
		// Legacy aliases
		apiGroup.GET("/orders", deps.ListOrders)
		apiGroup.POST("/orders", deps.CreateOrder)
		apiGroup.GET("/orders/:id/whatsapp", deps.OrderWhatsApp)
		apiGroup.GET("/orders/:id", deps.GetOrder)
		apiGroup.PATCH("/orders/:id", deps.PatchOrder)

		apiGroup.GET("/clients/template.xlsx", deps.ClientsTemplate)
		apiGroup.POST("/clients/import", deps.ImportClients)
		apiGroup.GET("/clients", deps.ListClients)
		apiGroup.POST("/clients", deps.CreateClient)
		apiGroup.PATCH("/clients/:id", deps.PatchClient)

		apiGroup.GET("/providers", deps.ListProviders)
		apiGroup.POST("/providers", deps.CreateProvider)
		apiGroup.GET("/technicians", deps.ListTechnicians)
		apiGroup.POST("/technicians", deps.CreateTechnician)
		apiGroup.GET("/products", deps.ListProducts)
		apiGroup.POST("/products", deps.CreateProduct)
		apiGroup.POST("/unit-types", deps.CreateUnitType)
		apiGroup.PATCH("/unit-types/:id", deps.PatchUnitType)
		apiGroup.POST("/product-types", deps.CreateProductType)
		apiGroup.PATCH("/product-types/:id", deps.PatchProductType)

		apiGroup.GET("/followups", deps.ListFollowups)
		apiGroup.PATCH("/followups/:id", deps.PatchFollowup)
		apiGroup.GET("/followup-rules", deps.ListFollowupRules)
		apiGroup.POST("/followup-rules", deps.CreateFollowupRule)

		apiGroup.GET("/users", deps.ListUsers)
		apiGroup.POST("/users", deps.CreateUser)

		apiGroup.GET("/whatsapp/templates", deps.ListWhatsAppTemplates)
		apiGroup.POST("/whatsapp/templates", deps.CreateWhatsAppTemplate)
		apiGroup.PATCH("/whatsapp/templates/:id", deps.PatchWhatsAppTemplate)
		apiGroup.GET("/tariffs", deps.ListTariffs)
		apiGroup.POST("/tariffs", deps.CreateTariff)
		apiGroup.GET("/results", deps.Results)
	}
}
