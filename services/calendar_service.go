package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

type CalendarService struct {
	config     *oauth2.Config
	tokens     map[string]*oauth2.Token
	tokenMutex sync.RWMutex
}

func NewCalendarService() (*CalendarService, error) {
	credentials := os.Getenv("GOOGLE_CREDENTIALS_JSON")
	if credentials == "" {
		// FIXME: Try to load from local file in development environment
		credBytes, err := os.ReadFile("./config/credentials.json")
		if err != nil {
			return nil, fmt.Errorf("GOOGLE_CREDENTIALS_JSON not set and couldn't read credentials.json: %v", err)
		}
		credentials = string(credBytes)
	}

	config, err := google.ConfigFromJSON([]byte(credentials), calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret: %v", err)
	}

	redirectURI := os.Getenv("GOOGLE_OAUTH_REDIRECT_URI")
	if redirectURI == "" {
		redirectURI = "http://localhost:8080/calendar/auth/callback" // fallback for development
	}
	config.RedirectURL = redirectURI

	return &CalendarService{
		config: config,
		tokens: make(map[string]*oauth2.Token),
	}, nil
}

func (s *CalendarService) GetAuthURL(state string) string {
	return s.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (s *CalendarService) HandleCallback(code string, userID string) error {
	token, err := s.config.Exchange(context.Background(), code)
	if err != nil {
		return fmt.Errorf("failed to exchange token: %v", err)
	}

	s.tokenMutex.Lock()
	s.tokens[userID] = token
	s.tokenMutex.Unlock()

	// In production, save token to persistent storage
	return s.saveToken(userID, token)
}

// saveToken saves the token to a file (in production, use a database)
func (s *CalendarService) saveToken(userID string, token *oauth2.Token) error {
	tokenPath := fmt.Sprintf("./config/token_%s.json", userID)
	f, err := os.OpenFile(tokenPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to create token file: %v", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

// GetCalendarService creates a Calendar service for a specific user
func (s *CalendarService) GetCalendarService(userID string) (*calendar.Service, error) {
	s.tokenMutex.RLock()
	token, exists := s.tokens[userID]
	s.tokenMutex.RUnlock()

	if !exists {
		// Try to load from file (in production, load from database)
		token, err := s.loadToken(userID)
		if err != nil {
			return nil, fmt.Errorf("no token found for user: %v", err)
		}
		s.tokenMutex.Lock()
		s.tokens[userID] = token
		s.tokenMutex.Unlock()
	}

	client := s.config.Client(context.Background(), token)
	return calendar.New(client)
}

// loadToken loads a token from file (in production, load from database)
func (s *CalendarService) loadToken(userID string) (*oauth2.Token, error) {
	tokenPath := fmt.Sprintf("./config/token_%s.json", userID)
	f, err := os.Open(tokenPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

// GetUpcomingEvents retrieves upcoming calendar events for a specific user
func (s *CalendarService) GetUpcomingEvents(userID string) ([]*calendar.Event, error) {
	srv, err := s.GetCalendarService(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get calendar service: %v", err)
	}

	t := time.Now().Format(time.RFC3339)
	events, err := srv.Events.List("primary").
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(t).
		MaxResults(10).
		OrderBy("startTime").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve events: %v", err)
	}

	return events.Items, nil
}
