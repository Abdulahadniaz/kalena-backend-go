package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

type CalendarService struct {
	config      *oauth2.Config
	redisClient *redis.Client
	ctx         context.Context
}

type UserSession struct {
	UserID       string    `json:"user_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
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

	// Parse Redis URL
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379" // Default local Redis URL
	}

	// Configure Redis client with Upstash settings
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %v", err)
	}

	// Add additional configuration for production
	opt.MaxRetries = 3
	opt.MinIdleConns = 2
	opt.PoolSize = 5
	opt.ConnMaxLifetime = time.Hour
	opt.PoolTimeout = 30 * time.Second

	redisClient := redis.NewClient(opt)

	// Test connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %v", err)
	}

	return &CalendarService{
		config:      config,
		redisClient: redisClient,
		ctx:         ctx,
	}, nil
}

func (s *CalendarService) GetAuthURL(state string) string {
	return s.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (s *CalendarService) HandleCallback(code string) (string, error) {
	fmt.Println("Handling callback")
	token, err := s.config.Exchange(s.ctx, code)
	if err != nil {
		return "", fmt.Errorf("failed to exchange token: %v", err)
	}

	userID := s.GenerateRandomUserID()

	userSession := UserSession{
		UserID:       userID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}

	sessionKey := "user_session:" + userID
	sessionJSON, _ := json.Marshal(userSession)
	err = s.redisClient.Set(context.Background(), sessionKey, sessionJSON, 24*time.Hour).Err()
	if err != nil {
		return "", fmt.Errorf("failed to save session to Redis: %v", err)
	}

	return s.saveToken(token, userID)
}

func (s *CalendarService) saveToken(token *oauth2.Token, userID string) (string, error) {
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token: %v", err)
	}

	key := fmt.Sprintf("oauth:token:%s", userID)
	err = s.redisClient.Set(s.ctx, key, tokenJSON, 24*time.Hour).Err()
	if err != nil {
		// Log the error details
		fmt.Printf("Redis save error: %v\n", err)
		return "", fmt.Errorf("failed to save token to Redis: %v", err)
	}

	return userID, nil
}

// GetCalendarService creates a Calendar service for a specific user
func (s *CalendarService) GetCalendarService(userID string) (*calendar.Service, error) {
	token, err := s.loadToken(userID)
	if err != nil {
		return nil, fmt.Errorf("no token found for user: %v", err)
	}

	client := s.config.Client(s.ctx, token)
	return calendar.New(client)
}

// loadToken loads a token from file (in production, load from database)
func (s *CalendarService) loadToken(userID string) (*oauth2.Token, error) {
	key := fmt.Sprintf("oauth:token:%s", userID)

	// Add retry logic for Redis operations
	var tokenJSON string
	var err error
	for retries := 0; retries < 3; retries++ {
		tokenJSON, err = s.redisClient.Get(s.ctx, key).Result()
		if err == nil {
			break
		}
		if err == redis.Nil {
			return nil, fmt.Errorf("token not found for user %s", userID)
		}
		time.Sleep(time.Duration(retries+1) * 100 * time.Millisecond)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load token from Redis after retries: %v", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal([]byte(tokenJSON), &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %v", err)
	}

	// Token refresh logic remains the same
	if token.Expiry.Before(time.Now()) {
		newToken, err := s.config.TokenSource(s.ctx, &token).Token()
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token: %v", err)
		}
		if _, err := s.saveToken(newToken, userID); err != nil {
			return nil, fmt.Errorf("failed to save refreshed token: %v", err)
		}
		return newToken, nil
	}

	return &token, nil
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

func (s *CalendarService) Close() error {
	return s.redisClient.Close()
}

func (s *CalendarService) GenerateRandomUserID() string {
	return uuid.New().String()
}

func (s *CalendarService) SaveStateToRedis(state string) {
	s.redisClient.Set(s.ctx, "oauth_state:"+state, "pending", 10*time.Minute)
}

func (s *CalendarService) GetStateFromRedis(state string) (string, error) {
	return s.redisClient.Get(s.ctx, "oauth_state:"+state).Result()
}

// ValidateToken checks if the token exists in Redis and returns the associated userID
func (cs *CalendarService) ValidateToken(token string) (string, error) {
	// Get userID from Redis using the token as key
	userID, err := cs.redisClient.Get(context.Background(), "token:"+token).Result()
	if err != nil {
		return "", err
	}
	return userID, nil
}
