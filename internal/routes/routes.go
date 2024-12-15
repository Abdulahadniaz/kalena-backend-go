package routes

import (
	"net/http"
	"time"

	"gin-server/internal/handlers"

	"github.com/gin-gonic/gin"
)

type Router struct {
	calendarController *handlers.CalendarController
}

func NewRouter(calendarController *handlers.CalendarController) *Router {
	return &Router{
		calendarController: calendarController,
	}
}

func (r *Router) SetupRoutes(router *gin.Engine) {
	// Entry point route
	router.GET("/", r.Home)

	// OAuth routes
	router.GET("/calendar/auth", r.calendarController.HandleGoogleAuth)
	router.GET("/calendar/auth/callback", r.calendarController.HandleGoogleCallback)

	// Protected calendar routes
	authorized := router.Group("/calendar")
	authorized.Use(r.calendarController.AuthMiddleware())
	{
		authorized.GET("/events", r.calendarController.GetUpcomingEvents)
	}

	// Optional: Add a health check route
	router.GET("/health", r.HealthCheck)
}

// Home is the entry point route
func (r *Router) Home(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message":       "Welcome to Google Calendar API Service",
		"documentation": "Contact admin for API documentation",
	})
}

// HealthCheck provides a simple endpoint to check if the service is running
func (r *Router) HealthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
	})
}
