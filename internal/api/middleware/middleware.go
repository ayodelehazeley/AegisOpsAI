// Package middleware provides HTTP middleware for the AegisOpsAI API server.
//
// Middleware functions follow the standard Chi middleware signature and can be
// composed into middleware chains using chi.Use.
package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/AegisOpsAI/aegisops/internal/auth"
	"github.com/AegisOpsAI/aegisops/internal/models"
)

// contextKey is an unexported type used for context value keys to prevent
// collisions with keys defined in other packages.
type contextKey string

const (
	// RequestIDKey is the context key for the unique request identifier.
	RequestIDKey contextKey = "requestID"
	// UserClaimsKey is the context key for the authenticated user's JWT claims.
	UserClaimsKey contextKey = "userClaims"
)

// RequestIDHeader is the HTTP header used to propagate the request ID.
const RequestIDHeader = "X-Request-ID"

// GetRequestID retrieves the request ID from the context.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// GetUserClaims retrieves the authenticated user's claims from the context.
func GetUserClaims(ctx context.Context) *auth.Claims {
	if claims, ok := ctx.Value(UserClaimsKey).(*auth.Claims); ok {
		return claims
	}
	return nil
}

// RequestID generates a unique ID for each request and stores it in the
// request context and response headers.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if id == "" {
			id = uuid.New().String()
		}

		ctx := context.WithValue(r.Context(), RequestIDKey, id)
		w.Header().Set(RequestIDHeader, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Logging logs incoming HTTP requests with structured fields including
// method, path, status code, duration, and request ID.
func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(ww, r)

			duration := time.Since(start)
			logger.Info("http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.statusCode),
				slog.Duration("duration", duration),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("request_id", GetRequestID(r.Context())),
				slog.String("user_agent", r.UserAgent()),
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

// WriteHeader captures the status code before delegating to the wrapped writer.
func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.statusCode = code
		rw.wroteHeader = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

// Write delegates to the wrapped writer, ensuring WriteHeader is called.
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// Recovery recovers from panics and returns a 500 Internal Server Error.
// It logs the panic stack trace for debugging.
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					stack := debug.Stack()
					logger.Error("panic recovered",
						slog.Any("panic", rec),
						slog.String("stack", string(stack)),
						slog.String("path", r.URL.Path),
						slog.String("request_id", GetRequestID(r.Context())),
					)
					http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// CORS returns middleware that sets Cross-Origin Resource Sharing headers.
// For production use, allowedOrigins should be configured explicitly.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	originsSet := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originsSet[o] = struct{}{}
	}
	allowAll := len(allowedOrigins) == 0

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if allowAll {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if _, ok := originsSet[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Request-ID")
			w.Header().Set("Access-Control-Max-Age", "3600")

			// Handle preflight requests.
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Auth returns middleware that validates JWT tokens from the Authorization
// header (Bearer scheme) and injects the parsed claims into the request context.
func Auth(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				http.Error(w, `{"error":"invalid authorization header format"}`, http.StatusUnauthorized)
				return
			}

			claims, err := authService.ValidateToken(parts[1])
			if err != nil {
				msg := fmt.Sprintf(`{"error":"%s"}`, err.Error())
				http.Error(w, msg, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RoleRequired returns middleware that restricts access to users with one of
// the specified roles. It must be applied after the Auth middleware.
func RoleRequired(roles ...models.Role) func(http.Handler) http.Handler {
	allowed := make(map[models.Role]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetUserClaims(r.Context())
			if claims == nil {
				http.Error(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
				return
			}

			if _, ok := allowed[claims.Role]; !ok {
				http.Error(w, `{"error":"insufficient permissions"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
