package main

import (
	"gin-server/middleware"
	"gin-server/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Create a default gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(middleware.CORSMiddleware())

	// Setup routes
	routes.SetupRoutes(router)

	// Run the server on port 8080
	router.Run(":8080")
}
