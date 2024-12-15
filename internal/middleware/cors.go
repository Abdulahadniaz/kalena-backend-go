package middleware

import (
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSConfig creates a CORS middleware configuration
func CORSConfig() gin.HandlerFunc {
	// Get frontend URL from environment variable
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000" // Default fallback
	}

	// Trim trailing slashes to ensure consistent matching
	frontendURL = strings.TrimRight(frontendURL, "/")

	return cors.New(cors.Config{
		// Allow specific origin
		AllowOrigins: []string{
			frontendURL,
			// Add additional allowed origins if needed
			"http://localhost:3000",
			"https://localhost:3000",
		},

		// Allow specific HTTP methods
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"HEAD",
			"OPTIONS",
		},

		// Allow specific headers
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"Accept",
		},

		// Expose specific headers to the client
		ExposeHeaders: []string{
			"Content-Length",
			"X-Total-Count",
		},

		// Allow credentials (cookies, authorization headers, etc.)
		AllowCredentials: true,

		// Cache preflight request results for 12 hours
		MaxAge: 12 * 60 * 60,
	})
}
