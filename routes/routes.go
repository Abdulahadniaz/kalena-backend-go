package routes

import (
	"gin-server/controllers"
	"gin-server/middleware"
	"gin-server/services"
	"log"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	router.Use(middleware.CORSMiddleware())
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

	calendar := router.Group("/calendar")
	{
		calendar.GET("/auth", calendarController.InitiateGoogleAuth)
		calendar.GET("/auth/callback", calendarController.HandleGoogleCallback)

		// Protected routes
		protected := calendar.Group("")
		protected.Use(calendarController.AuthMiddleware())
		{
			calendar.GET("/upcoming-events", calendarController.GetUpcomingEvents)
			// Add other protected routes here
		}
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

	// Protected routes
	protected := router.Group("/")
	protected.Use(calendarController.AuthMiddleware())
	{
		protected.GET("/upcoming-events", calendarController.GetUpcomingEvents)
		// Add other protected routes here
	}
}
