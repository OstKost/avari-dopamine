package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ostkost/dopamine-market/api/internal/modules/identity/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/identity/usecase"
	"github.com/ostkost/dopamine-market/api/internal/platform/httpserver"
)

type Handler struct {
	authUC          *usecase.AuthUseCase
	refreshTokenTTL time.Duration
	accessTokenTTL  time.Duration
	isSecureCookie  bool
}

func NewHandler(
	authUC *usecase.AuthUseCase,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
	isSecureCookie bool,
) *Handler {
	return &Handler{
		authUC:          authUC,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		isSecureCookie:  isSecureCookie,
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/register", h.handleRegister)
	r.Post("/login", h.handleLogin)
	r.Post("/refresh", h.handleRefresh)
	r.Post("/logout", h.handleLogout)

	return r
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

type AuthResponse struct {
	User        UserResponse `json:"user"`
	AccessToken string       `json:"access_token"`
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.authUC.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidEmail):
			h.writeError(w, http.StatusBadRequest, "invalid email format")
		case errors.Is(err, domain.ErrWeakPassword):
			h.writeError(w, http.StatusBadRequest, "password must be at least 8 characters long")
		case errors.Is(err, domain.ErrEmailAlreadyExists):
			h.writeError(w, http.StatusConflict, "user with this email already exists")
		default:
			h.writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	h.setAuthCookies(w, resp.Tokens)

	h.writeJSON(w, http.StatusCreated, AuthResponse{
		User: UserResponse{
			ID:        resp.User.ID().String(),
			Email:     resp.User.Email().String(),
			CreatedAt: resp.User.CreatedAt().Format(time.RFC3339),
		},
		AccessToken: resp.Tokens.AccessToken,
	})
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.authUC.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			h.writeError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.setAuthCookies(w, resp.Tokens)

	h.writeJSON(w, http.StatusOK, AuthResponse{
		User: UserResponse{
			ID:        resp.User.ID().String(),
			Email:     resp.User.Email().String(),
			CreatedAt: resp.User.CreatedAt().Format(time.RFC3339),
		},
		AccessToken: resp.Tokens.AccessToken,
	})
}

func (h *Handler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken := ""
	if cookie, err := r.Cookie(httpserver.RefreshTokenCookie); err == nil {
		refreshToken = cookie.Value
	}

	if refreshToken == "" {
		// Fallback: попытаться прочитать из body
		var bodyReq struct {
			RefreshToken string `json:"refresh_token"`
		}
		_ = json.NewDecoder(r.Body).Decode(&bodyReq)
		refreshToken = bodyReq.RefreshToken
	}

	if refreshToken == "" {
		h.writeError(w, http.StatusUnauthorized, "missing refresh token")
		return
	}

	tokens, err := h.authUC.Refresh(r.Context(), refreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrRefreshTokenReused) {
			h.clearAuthCookies(w)
			h.writeError(w, http.StatusUnauthorized, "refresh token reuse detected: session terminated")
			return
		}
		h.writeError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	h.setAuthCookies(w, *tokens)

	h.writeJSON(w, http.StatusOK, map[string]string{
		"access_token": tokens.AccessToken,
	})
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	refreshToken := ""
	if cookie, err := r.Cookie(httpserver.RefreshTokenCookie); err == nil {
		refreshToken = cookie.Value
	}

	if refreshToken != "" {
		_ = h.authUC.Logout(r.Context(), refreshToken)
	}

	h.clearAuthCookies(w)
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) HandleMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpserver.UserIDFromContext(r.Context())
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.authUC.GetUserByID(r.Context(), userID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "user not found")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"user": UserResponse{
			ID:        user.ID().String(),
			Email:     user.Email().String(),
			CreatedAt: user.CreatedAt().Format(time.RFC3339),
		},
	})
}

func (h *Handler) setAuthCookies(w http.ResponseWriter, tokens usecase.TokenPair) {
	http.SetCookie(w, &http.Cookie{
		Name:     httpserver.AccessTokenCookie,
		Value:    tokens.AccessToken,
		Path:     "/",
		Expires:  time.Now().Add(h.accessTokenTTL),
		HttpOnly: true,
		Secure:   h.isSecureCookie,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     httpserver.RefreshTokenCookie,
		Value:    tokens.RefreshToken,
		Path:     "/auth",
		Expires:  time.Now().Add(h.refreshTokenTTL),
		HttpOnly: true,
		Secure:   h.isSecureCookie,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) clearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     httpserver.AccessTokenCookie,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     httpserver.RefreshTokenCookie,
		Value:    "",
		Path:     "/auth",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
	})
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, msg string) {
	h.writeJSON(w, status, map[string]string{
		"error":   http.StatusText(status),
		"message": msg,
	})
}
