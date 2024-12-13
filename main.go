package main

import (
	"github.com/gin-gonic/gin"
	"gin-server/routes"
	"gin-server/middleware"
	"fmt"
)

func main() {
    // Create a default gin router
    router := gin.Default()

    // Add CORS middleware
    router.Use(middleware.CORSMiddleware())

    // Setup routes
    routes.SetupRoutes(router)

    // Print out all registered routes (for debugging)
    for _, route := range router.Routes() {
        fmt.Printf("Method: %v, Path: %v\n", route.Method, route.Path)
    }

    // Run the server on port 8080
    router.Run(":8080")
}