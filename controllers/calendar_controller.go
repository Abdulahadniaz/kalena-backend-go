package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"gin-server/services"
	"net/http"

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

	// In production, store state in session/redis with expiry
	// For now, we'll use a cookie
	c.SetCookie("oauth_state", state, 3600, "/", "", false, true)

	// Get the authorization URL
	authURL := cc.calendarService.GetAuthURL(state)

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
	})
}

// HandleGoogleCallback processes the OAuth callback
func (cc *CalendarController) HandleGoogleCallback(c *gin.Context) {
	// Get state and code from query params
	state := c.Query("state")
	code := c.Query("code")

	// Get stored state from cookie
	storedState, err := c.Cookie("oauth_state")
	if err != nil || state != storedState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}

	// Clear the state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	// Get user ID from session/token
	// For now, using a placeholder. In production, get this from your auth system
	userID := "test_user"

	// Exchange code for token and store it
	err = cc.calendarService.HandleCallback(code, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to handle callback"})
		return
	}

	// Redirect to frontend or return success
	// For now, returning success message
	c.JSON(http.StatusOK, gin.H{
		"message": "Calendar successfully connected",
	})
}

// GetUpcomingEvents handles the request for upcoming calendar events
func (cc *CalendarController) GetUpcomingEvents(c *gin.Context) {
	// In production, get userID from authenticated session
	userID := "test_user"

	events, err := cc.calendarService.GetUpcomingEvents(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
	})
}
