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

	router.GET("/", r.Home)

	router.GET("/calendar/auth", r.calendarController.HandleGoogleAuth)
	router.GET("/calendar/auth/callback", r.calendarController.HandleGoogleCallback)

	authorized := router.Group("/calendar")
	authorized.Use(r.calendarController.AuthMiddleware())
	{
		authorized.GET("/events", r.calendarController.GetUpcomingEvents)
	}

	router.GET("/health", r.HealthCheck)
}

func (r *Router) Home(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message":       "Welcome to Google Calendar API Service",
		"documentation": "Contact admin for API documentation",
	})
}

func (r *Router) HealthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
	})
}
