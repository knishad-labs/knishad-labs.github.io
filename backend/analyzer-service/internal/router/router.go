package router

import (
	"analyzer-service/internal/controller"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures endpoints for the analyzer service
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
		api.POST("/analyze", controller.AnalyzeQuery)
		api.GET("/reports", controller.GetReports)
		api.GET("/reports/:id", controller.GetReport)
		api.POST("/reports/:id/apply", controller.ApplyFix)
	}

	return r
}
