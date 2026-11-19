// Package auth provides JWT-based authentication and password hashing
// for the AegisOpsAI platform.
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/AegisOpsAI/aegisops/internal/models"
)

// Standard authentication errors.
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrTokenExpired       = errors.New("token has expired")
	ErrInsufficientRole   = errors.New("insufficient role permissions")
)

// Config holds the configuration for the auth service.
type Config struct {
	// JWTSecret is the HMAC signing key for tokens.
	JWTSecret string
	// AccessTokenTTL is the lifetime of an access token.
	AccessTokenTTL time.Duration
	// RefreshTokenTTL is the lifetime of a refresh token.
	RefreshTokenTTL time.Duration
	// Issuer is the JWT issuer claim value.
	Issuer string
}

// DefaultConfig returns a Config with sensible defaults.
// In production, JWTSecret must be overridden with a strong secret.
func DefaultConfig() Config {
	return Config{
		JWTSecret:       "change-me-in-production",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "aegisops",
	}
}

// Claims represents the JWT claims payload for AegisOpsAI tokens.
type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID   `json:"uid"`
	Email  string      `json:"email"`
	Role   models.Role `json:"role"`
}

// TokenPair holds an access and refresh token pair.
type TokenPair struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

// Service provides authentication operations.
type Service struct {
	cfg Config
}

// NewService creates a new auth Service with the given configuration.
func NewService(cfg Config) *Service {
	return &Service{cfg: cfg}
}

// HashPassword generates a bcrypt hash for the given plaintext password.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}
	return string(hash), nil
}

// CheckPassword compares a plaintext password against a bcrypt hash.
// Returns nil on match, or ErrInvalidCredentials on mismatch.
func CheckPassword(hash, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return ErrInvalidCredentials
	}
	return nil
}

// GenerateTokenPair creates a new access/refresh token pair for the given user.
func (s *Service) GenerateTokenPair(user *models.User) (*TokenPair, error) {
	now := time.Now()
	accessExpiry := now.Add(s.cfg.AccessTokenTTL)

	// Build access token claims.
	accessClaims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.cfg.Issuer,
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			ID:        uuid.New().String(),
		},
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
	}

	accessToken, err := s.signToken(accessClaims)
	if err != nil {
		return nil, fmt.Errorf("signing access token: %w", err)
	}

	// Build refresh token claims (minimal payload).
	refreshClaims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.cfg.Issuer,
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.RefreshTokenTTL)),
			ID:        uuid.New().String(),
		},
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
	}

	refreshToken, err := s.signToken(refreshClaims)
	if err != nil {
		return nil, fmt.Errorf("signing refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiry,
	}, nil
}

// ValidateToken parses and validates a JWT token string, returning the claims.
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		// Ensure signing method is HMAC.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshTokens validates a refresh token and issues a new token pair.
// The user argument should be the current user record from the database,
// ensuring that role or email changes are reflected in the new tokens.
func (s *Service) RefreshTokens(refreshTokenString string, user *models.User) (*TokenPair, error) {
	claims, err := s.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("validating refresh token: %w", err)
	}

	// Ensure the refresh token belongs to the claimed user.
	if claims.UserID != user.ID {
		return nil, ErrInvalidToken
	}

	return s.GenerateTokenPair(user)
}

// signToken signs the given claims into a JWT string using HS256.
func (s *Service) signToken(claims Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}
