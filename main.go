package main

import (
	config "gin-server/internal/db"
	"gin-server/internal/handlers"
	"gin-server/internal/middleware"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	config.ConnectDatabase()

	r := gin.Default()
	// use cors middleware defined in middleware/cors.go
	r.Use(middleware.CorsMiddleware())

	auth := r.Group("/auth")
	{
		authHandler := handlers.NewAuthHandler(config.DB)
		auth.POST("/signup", authHandler.Signup)
		auth.POST("/login", authHandler.Login)
	}

	// protected := r.Group("/api")
	// protected.Use(middleware.AuthMiddleware())
	// {
	// 	protected.GET("/profile", func(c *gin.Context) {
	// 		userID, _ := c.Get("user_id")
	// 		email, _ := c.Get("email")
	// 		c.JSON(200, gin.H{
	// 			"user_id": userID,
	// 			"email":   email,
	// 		})
	// 	})
	// }

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
