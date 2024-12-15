package main

import (
	"log"
	"os"

	config "gin-server/configs"
	"gin-server/internal/handlers"
	"gin-server/internal/middleware"
	"gin-server/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load environment (you might want to use a package like godotenv in a real app)
	gin.SetMode(getGinMode())

	// Create a new Gin engine
	r := gin.Default()

	// Apply CORS middleware
	r.Use(middleware.CORSConfig())

	// Initialize Calendar Controller
	calendarController, err := handlers.NewCalendarController(config.LoadConfig().OAuthCredentialsPath)
	if err != nil {
		log.Fatalf("Failed to create calendar controller: %v", err)
	}

	// Create router with routes
	router := routes.NewRouter(calendarController)

	// Setup routes
	router.SetupRoutes(r)

	// Start server
	port := getServerPort()
	log.Printf("Server starting on %s", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Helper function to get Gin mode from environment
func getGinMode() string {
	mode := os.Getenv("GIN_MODE")
	switch mode {
	case "release":
		return gin.ReleaseMode
	case "test":
		return gin.TestMode
	default:
		return gin.DebugMode
	}
}

// Helper function to get server port from environment
func getServerPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return ":8080"
	}
	return ":" + port
}
