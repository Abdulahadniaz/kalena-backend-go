package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

type Service struct {
	config *oauth2.Config
}

func NewOAuthService(credentialsPath string) (*Service, error) {
	googleCreds := os.Getenv("GOOGLE_CREDENTIALS_JSON")
	if googleCreds == "" {
		b, err := os.ReadFile(credentialsPath)
		if err != nil {
			return nil, fmt.Errorf("unable to read client secret file: %v", err)
		}
		googleCreds = string(b)
	}

	config, err := google.ConfigFromJSON([]byte(googleCreds), calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	redirectURI := os.Getenv("GOOGLE_OAUTH_REDIRECT_URI")
	if redirectURI == "" {
		redirectURI = "http://localhost:8080/calendar/auth/callback"
	}

	config.RedirectURL = redirectURI

	return &Service{config: config}, nil
}

func (s *Service) GetAuthURL() string {
	return s.config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
}

func (s *Service) ExchangeToken(authCode string) (*oauth2.Token, error) {
	tok, err := s.config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %v", err)
	}
	return tok, nil
}

func (s *Service) GetClient(tok *oauth2.Token) *http.Client {
	return s.config.Client(context.Background(), tok)
}

func (s *Service) SaveToken(path string, token *oauth2.Token) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %v", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

func (s *Service) LoadToken(path string) (*oauth2.Token, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}
