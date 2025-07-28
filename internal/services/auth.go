package services

import (
	"context"
	"fmt"
	"github.com/SAP-2025/auth-service/internal/config"
	"github.com/SAP-2025/auth-service/internal/utils"
	"github.com/google/uuid"
	"net/http"
	"time"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"golang.org/x/oauth2"
)

type AuthService struct {
	cfg           *config.Config
	pkceStore     *PKCEStore
	casdoorClient *casdoorsdk.Client
	oauth2Config  *oauth2.Config
}

func NewAuthService(pkceStore *PKCEStore, cfg *config.Config) *AuthService {
	client := config.NewCasdoorClient(cfg)

	oauth2Config := &oauth2.Config{
		ClientID:     cfg.OAuth2.Casdoor.ClientID,
		ClientSecret: cfg.OAuth2.Casdoor.ClientSecret,
		RedirectURL:  cfg.OAuth2.Casdoor.RedirectURI,
		Scopes:       []string{"read"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.OAuth2.Casdoor.BaseURL + "/api/login/oauth/authorize",
			TokenURL: cfg.OAuth2.Casdoor.BaseURL + "/api/login/oauth/access_token",
		},
	}

	return &AuthService{
		casdoorClient: client,
		oauth2Config:  oauth2Config,
		pkceStore:     pkceStore,
	}
}

type LoginResponse struct {
	LoginURL  string `json:"login_url"`
	SessionID string `json:"session_id"`
}

type CallbackResponse struct {
	AccessToken  string             `json:"access_token"`
	RefreshToken string             `json:"refresh_token"`
	ExpiresIn    int64              `json:"expires_in"`
	User         *casdoorsdk.Claims `json:"user"`
}

func (s *AuthService) GetLoginURL() (*LoginResponse, error) {
	sessionID := uuid.New().String()
	pkceChallenge := utils.NewPKCEChallenge()

	err := s.pkceStore.SavePKCE(sessionID, pkceChallenge)
	if err != nil {
		return nil, fmt.Errorf("failed to save PKCE: %w", err)
	}

	authURL := s.oauth2Config.AuthCodeURL(sessionID,
		oauth2.SetAuthURLParam("code_challenge", pkceChallenge.CodeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", pkceChallenge.Method),
	)

	return &LoginResponse{
		LoginURL:  authURL,
		SessionID: sessionID,
	}, nil
}

func (s *AuthService) ExchangeCode(code, state string) (*CallbackResponse, error) {
	pkceChallenge, err := s.pkceStore.GetAndDeletePKCE(state)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired session: %w", err)
	}

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{
		Timeout: 10 * time.Second,
	})

	token, err := s.oauth2Config.Exchange(ctx, code,
		oauth2.SetAuthURLParam("code_verifier", pkceChallenge.CodeVerifier),
	)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	user, err := s.casdoorClient.ParseJwtToken(token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	return &CallbackResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    token.Expiry.Unix(),
		User:         user,
	}, nil
}

func (s *AuthService) ParseUser(accessToken string) (*casdoorsdk.Claims, error) {
	return s.casdoorClient.ParseJwtToken(accessToken)
}

func (s *AuthService) ValidateSession(sessionID string) bool {
	return s.pkceStore.ExistsPKCE(sessionID)
}

func (s *AuthService) CancelSession(sessionID string) error {
	return s.pkceStore.DeletePKCE(sessionID)
}
