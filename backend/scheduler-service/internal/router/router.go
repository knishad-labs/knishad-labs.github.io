package router

import (
	"scheduler-service/internal/controller"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures endpoints for the scheduler service
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
		api.GET("/tasks", controller.GetTasks)
		api.GET("/tasks/:id", controller.GetTask)
		api.POST("/tasks", controller.CreateTask)
		api.PUT("/tasks/:id", controller.UpdateTask)
		api.DELETE("/tasks/:id", controller.DeleteTask)
	}

	return r
}
