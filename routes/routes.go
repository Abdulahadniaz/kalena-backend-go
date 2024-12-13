package routes

import (
    "gin-server/controllers"
	"gin-server/services"
    "github.com/gin-gonic/gin"
	"log"
)

func SetupRoutes(router *gin.Engine) {
    calendarService, err := services.NewCalendarService()
	
	if err != nil {
		log.Fatalf("Failed to create calendar service: %v", err)
	}

    userController := controllers.NewUserController()
	calendarController := controllers.NewCalendarController(calendarService)


    // User routes
    userRoutes := router.Group("/user")
    {
        userRoutes.POST("", userController.CreateUser)
        userRoutes.GET("", userController.GetUser)
    }

	calendarRoutes := router.Group("/calendar")
	{
		calendarRoutes.GET("/upcoming-events", calendarController.GetUpcomingEvents)
		calendarRoutes.GET("/auth", calendarController.InitiateGoogleAuth)
        calendarRoutes.GET("/auth/callback", calendarController.HandleGoogleCallback)
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