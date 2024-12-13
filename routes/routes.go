package routes

import (
    "gin-server/controllers"
    "github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
    userController := controllers.NewUserController()

    // User routes
    userRoutes := router.Group("/user")
    {
        userRoutes.POST("", userController.CreateUser)
        userRoutes.GET("", userController.GetUser)
    }

    // Home route
    router.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "Welcome to Gin Server!",
        })
    })

    // Ping route
    router.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "pong",
        })
    })
}