package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"gin-server/services"
	"net/http"
	"os"

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

	c.SetCookie("oauth_state", state, 3600, "/", "localhost", false, true)
	fmt.Printf("Setting new state: %s\n", state)

	authURL := cc.calendarService.GetAuthURL(state)
	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
	})
}

// HandleGoogleCallback processes the OAuth callback
func (cc *CalendarController) HandleGoogleCallback(c *gin.Context) {
	code := c.Query("code")

	// Get user ID from session/token
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

	// Exchange code for token and store it
	err := cc.calendarService.HandleCallback(code, userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to handle callback"})
		return
	}

	// Redirect to frontend
	c.Redirect(http.StatusFound, os.Getenv("FRONTEND_URL"))
	c.JSON(http.StatusOK, gin.H{
		"message": "Calendar successfully connected",
	})
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
