package router

import (
	"config-service/internal/controller"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures endpoints for the config service
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	api := r.Group("/api")
	{
		// Database Connection CRUD
		api.GET("/db-connections", controller.GetDatabaseConnections)
		api.GET("/db-connections/:id", controller.GetDatabaseConnection)
		api.POST("/db-connections", controller.CreateDatabaseConnection)
		api.PUT("/db-connections/:id", controller.UpdateDatabaseConnection)
		api.DELETE("/db-connections/:id", controller.DeleteDatabaseConnection)

		// AI Provider CRUD
		api.GET("/ai-providers", controller.GetAIProviders)
		api.GET("/ai-providers/:id", controller.GetAIProvider)
		api.POST("/ai-providers", controller.CreateAIProvider)
		api.PUT("/ai-providers/:id", controller.UpdateAIProvider)
		api.DELETE("/ai-providers/:id", controller.DeleteAIProvider)
	}

	return r
}
