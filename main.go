package main

import (
	"fmt"
	"gin-server/middleware"
	"gin-server/routes"
	"gin-server/services"

	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
)

func main() {
	// Create a default gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(middleware.CORSMiddleware())

	calendarService, err := services.NewCalendarService()
	if err != nil {
		log.Fatalf("Failed to create calendar service: %v", err)
	}

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nShutting down...")
		if err := calendarService.Close(); err != nil {
			log.Printf("Error closing Redis connection: %v", err)
		}
		os.Exit(0)
	}()

	// Setup routes
	routes.SetupRoutes(router)

	// Run the server on port 8080
	router.Run(":8080")
}
