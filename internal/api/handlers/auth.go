package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AegisOpsAI/aegisops/internal/auth"
	"github.com/AegisOpsAI/aegisops/internal/models"
)

// AuthHandler handles user registration, login, and token refresh.
type AuthHandler struct {
	db     *pgxpool.Pool
	auth   *auth.Service
	logger *slog.Logger
}

// NewAuthHandler creates an AuthHandler instance.
func NewAuthHandler(db *pgxpool.Pool, authSvc *auth.Service, logger *slog.Logger) *AuthHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &AuthHandler{
		db:     db,
		auth:   authSvc,
		logger: logger.With(slog.String("component", "auth-handler")),
	}
}

// LoginRequest is the JSON payload for user authentication.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest is the JSON payload for new user registration.
type RegisterRequest struct {
	Email       string      `json:"email"`
	Password    string      `json:"password"`
	DisplayName string      `json:"displayName"`
	Role        models.Role `json:"role,omitempty"`
}

// RefreshRequest is the JSON payload for token refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// Login authenticates a user and returns a JWT token pair.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var user models.User
	query := `
		SELECT id, email, password_hash, display_name, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	err := h.db.QueryRow(r.Context(), query, req.Email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			JSONError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		h.logger.Error("database error on user lookup", slog.String("error", err.Error()))
		JSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := auth.CheckPassword(user.PasswordHash, req.Password); err != nil {
		JSONError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	tokens, err := h.auth.GenerateTokenPair(&user)
	if err != nil {
		h.logger.Error("failed to generate token pair", slog.String("error", err.Error()))
		JSONError(w, http.StatusInternalServerError, "failed to generate tokens")
		return
	}

	JSON(w, http.StatusOK, map[string]any{
		"user": map[string]any{
			"id":          user.ID,
			"email":       user.Email,
			"displayName": user.DisplayName,
			"role":        user.Role,
		},
		"tokens": tokens,
	})
}

// Register provisions a new user record.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.DisplayName == "" {
		JSONError(w, http.StatusBadRequest, "email, password, and displayName are required")
		return
	}

	role := req.Role
	if role == "" {
		role = models.RoleOperator
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to secure password")
		return
	}

	newUser := models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: hash,
		DisplayName:  req.DisplayName,
		Role:         role,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	query := `
		INSERT INTO users (id, email, password_hash, display_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = h.db.Exec(r.Context(), query, newUser.ID, newUser.Email, newUser.PasswordHash, newUser.DisplayName, newUser.Role, newUser.CreatedAt, newUser.UpdatedAt)
	if err != nil {
		h.logger.Error("failed to create user", slog.String("error", err.Error()))
		JSONError(w, http.StatusConflict, "user with this email already exists")
		return
	}

	tokens, err := h.auth.GenerateTokenPair(&newUser)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to generate auth tokens")
		return
	}

	JSON(w, http.StatusCreated, map[string]any{
		"user": map[string]any{
			"id":          newUser.ID,
			"email":       newUser.Email,
			"displayName": newUser.DisplayName,
			"role":        newUser.Role,
		},
		"tokens": tokens,
	})
}

// Refresh handles token rotation via valid refresh token.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	claims, err := h.auth.ValidateToken(req.RefreshToken)
	if err != nil {
		JSONError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	user := models.User{
		ID:    claims.UserID,
		Email: claims.Email,
		Role:  claims.Role,
	}

	tokens, err := h.auth.GenerateTokenPair(&user)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to generate tokens")
		return
	}

	JSON(w, http.StatusOK, tokens)
}
