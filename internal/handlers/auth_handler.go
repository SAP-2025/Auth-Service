package handlers

import (
	"encoding/json"
	"github.com/SAP-2025/auth-service/internal/services"
	"log"
	"net/http"
	"strings"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type SessionStatusResponse struct {
	Valid     bool   `json:"valid"`
	SessionID string `json:"session_id,omitempty"`
}

type ProfileResponse struct {
	User interface{} `json:"user"`
}

// writeJSON helper function
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError helper function
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}

// getCookie helper function
func getCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// setCookie helper function
func setCookie(w http.ResponseWriter, name, value string, maxAge int) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)
}

// Login handler
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	loginResp, err := h.authService.GetLoginURL()
	if err != nil {
		log.Printf("Login error: %v", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Set session cookie
	setCookie(w, "session_id", loginResp.SessionID, 600)

	writeJSON(w, http.StatusOK, loginResp)
}

// Callback handler
func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" || state == "" {
		writeError(w, http.StatusBadRequest, "Missing code or state")
		return
	}

	// Verify session cookie matches state
	sessionID, err := getCookie(r, "session_id")
	if err != nil || sessionID != state {
		writeError(w, http.StatusBadRequest, "Invalid session")
		return
	}

	// Exchange code for token
	callbackResp, err := h.authService.ExchangeCode(code, state)
	if err != nil {
		log.Printf("Callback error: %v", err)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Clear session cookie
	setCookie(w, "session_id", "", -1)

	writeJSON(w, http.StatusOK, callbackResp)
}

// Cancel login session
func (h *AuthHandler) CancelLogin(w http.ResponseWriter, r *http.Request) {
	sessionID, err := getCookie(r, "session_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "No active session")
		return
	}

	err = h.authService.CancelSession(sessionID)
	if err != nil {
		log.Printf("Cancel session error: %v", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Clear cookie
	setCookie(w, "session_id", "", -1)

	writeJSON(w, http.StatusOK, MessageResponse{Message: "Session cancelled"})
}

// Check session status
func (h *AuthHandler) SessionStatus(w http.ResponseWriter, r *http.Request) {
	sessionID, err := getCookie(r, "session_id")
	if err != nil {
		writeJSON(w, http.StatusOK, SessionStatusResponse{Valid: false})
		return
	}

	valid := h.authService.ValidateSession(sessionID)
	writeJSON(w, http.StatusOK, SessionStatusResponse{
		Valid:     valid,
		SessionID: sessionID,
	})
}

// Protected endpoint
func (h *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || len(authHeader) < 8 {
		writeError(w, http.StatusUnauthorized, "Missing or invalid authorization header")
		return
	}

	// Extract token
	if !strings.HasPrefix(authHeader, "Bearer ") {
		writeError(w, http.StatusUnauthorized, "Invalid authorization header format")
		return
	}

	token := authHeader[7:] // Remove "Bearer "

	// Parse user
	user, err := h.authService.ParseUser(token)
	if err != nil {
		log.Printf("Profile error: %v", err)
		writeError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	writeJSON(w, http.StatusOK, ProfileResponse{User: user})
}
