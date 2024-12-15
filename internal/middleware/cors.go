package middleware

import (
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSConfig() gin.HandlerFunc {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	frontendURL = strings.TrimRight(frontendURL, "/")

	return cors.New(cors.Config{
		AllowOrigins: []string{
			frontendURL,
			"http://localhost:3000",
			"https://localhost:3000",
		},

		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"HEAD",
			"OPTIONS",
		},

		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"Accept",
		},

		ExposeHeaders: []string{
			"Content-Length",
			"X-Total-Count",
		},

		AllowCredentials: true,

		MaxAge: 12 * 60 * 60,
	})
}
