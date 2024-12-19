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
	authCode := ctx.Query("code")
	if authCode == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code is missing"})
		return
	}

	tok, err := c.oauthService.ExchangeToken(authCode)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = c.oauthService.SaveToken("token.json", tok)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Redirect(http.StatusFound, os.Getenv("FRONTEND_URL"))
}

func (c *CalendarController) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		tok, err := c.oauthService.LoadToken("")
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			ctx.Abort()
			return
		}

		client := c.oauthService.GetClient(tok)

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
