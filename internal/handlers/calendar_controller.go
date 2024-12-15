package handlers

import (
	"gin-server/internal/calendar"
	"gin-server/internal/oauth"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CalendarController struct {
	oauthService    *oauth.Service
	calendarService *calendar.Service
}

func NewCalendarController(credentialsPath string) (*CalendarController, error) {
	// Initialize OAuth Service
	oauthService, err := oauth.NewOAuthService(credentialsPath)
	if err != nil {
		return nil, err
	}

	return &CalendarController{
		oauthService: oauthService,
	}, nil
}

func (c *CalendarController) HandleGoogleAuth(ctx *gin.Context) {
	authURL := c.oauthService.GetAuthURL()
	ctx.JSON(http.StatusOK, gin.H{"auth_url": authURL})
}

func (c *CalendarController) HandleGoogleCallback(ctx *gin.Context) {
	// Get authorization code from query parameters
	authCode := ctx.Query("code")
	if authCode == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code is missing"})
		return
	}

	// Exchange authorization code for token
	tok, err := c.oauthService.ExchangeToken(authCode)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Save token (you might want to save this in a database in a real-world scenario)
	err = c.oauthService.SaveToken("token.json", tok)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Redirect(http.StatusFound, os.Getenv("FRONTEND_URL"))
}

func (c *CalendarController) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Load token
		tok, err := c.oauthService.LoadToken("token.json")
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			ctx.Abort()
			return
		}

		// Create HTTP client with the token
		client := c.oauthService.GetClient(tok)

		// Initialize calendar service
		calendarService, err := calendar.NewCalendarService(client)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create calendar service"})
			ctx.Abort()
			return
		}
		c.calendarService = calendarService

		ctx.Next()
	}
}

func (c *CalendarController) GetUpcomingEvents(ctx *gin.Context) {
	// Get max results from query parameter, default to 10
	maxResults := 10
	if max, exists := ctx.GetQuery("max"); exists {
		if parsedMax, err := strconv.Atoi(max); err == nil {
			maxResults = parsedMax
		}
	}

	events, err := c.calendarService.GetUpcomingEvents(maxResults)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, events)
}
