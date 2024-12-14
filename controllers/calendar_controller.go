package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"gin-server/services"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type CalendarController struct {
	calendarService *services.CalendarService
}

func NewCalendarController(cs *services.CalendarService) *CalendarController {
	return &CalendarController{calendarService: cs}
}

// generateState creates a random state string for OAuth security
func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// InitiateGoogleAuth starts the OAuth flow
func (cc *CalendarController) InitiateGoogleAuth(c *gin.Context) {
	state, err := generateState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate state"})
		return
	}

	// store the state in redis with expiry time of 1 day
	cc.calendarService.SaveStateToRedis(state)

	authURL := cc.calendarService.GetAuthURL(state)
	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
	})

}

// HandleGoogleCallback processes the OAuth callback
func (cc *CalendarController) HandleGoogleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	// get the state from redis
	stateCheck, err := cc.calendarService.GetStateFromRedis(state)
	if err != nil || stateCheck != "pending" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get state from redis"})
		return
	}

	// Exchange code for token and store it
	userID, err := cc.calendarService.HandleCallback(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to handle callback"})
		return
	}

	// Make sure the FRONTEND_URL doesn't end with a trailing slash
	frontendURL := strings.TrimRight(os.Getenv("FRONTEND_URL"), "/")
	redirectURL := frontendURL + "/calendar/auth?user_id=" + userID + "&status=success"
	c.Redirect(http.StatusFound, redirectURL)
}

// GetUpcomingEvents handles the request for upcoming calendar events
func (cc *CalendarController) GetUpcomingEvents(c *gin.Context) {
	// Get userID from the session
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Convert interface{} to string
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	events, err := cc.calendarService.GetUpcomingEvents(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
	})
}

// AuthMiddleware validates the user's authentication status using Redis
func (cc *CalendarController) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No authorization header"})
			c.Abort()
			return
		}

		// Extract the token from the Authorization header
		// Expecting: "Bearer <token>"
		if len(authHeader) <= 7 || authHeader[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}
		token := authHeader[7:]

		// Validate token in Redis
		userID, err := cc.calendarService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set userID in the context for later use
		c.Set("userID", userID)
		c.Next()
	}
}
