package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/candlekeep/zot-artifact-store/internal/models"
	"zotregistry.io/zot/pkg/log"
)

// Middleware provides authentication and authorization middleware
type Middleware struct {
	jwtValidator *JWTValidator
	policyEngine *PolicyEngine
	logger       log.Logger
	enabled      bool
}

// NewMiddleware creates a new auth middleware
func NewMiddleware(jwtValidator *JWTValidator, policyEngine *PolicyEngine, logger log.Logger, enabled bool) *Middleware {
	return &Middleware{
		jwtValidator: jwtValidator,
		policyEngine: policyEngine,
		logger:       logger,
		enabled:      enabled,
	}
}

// AuthenticateRequest is middleware that validates JWT tokens
func (m *Middleware) AuthenticateRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication if disabled
		if !m.enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Try to extract bearer token
		token, err := ExtractBearerToken(r)
		if err != nil {
			// No token - check if anonymous access is allowed
			authCtx := &models.AuthContext{
				User:        nil,
				Permissions: m.policyEngine.GetPermissions(nil),
				IsAnonymous: true,
			}

			ctx := context.WithValue(r.Context(), AuthContextKey, authCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Validate token
		user, err := m.jwtValidator.ValidateToken(token)
		if err != nil {
			m.logger.Error().Err(err).Msg("failed to validate token")
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Get user permissions
		permissions := m.policyEngine.GetPermissions(user)

		// Create auth context
		authCtx := &models.AuthContext{
			User:        user,
			Permissions: permissions,
			IsAnonymous: false,
		}

		// Add user and auth context to request context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		ctx = context.WithValue(ctx, AuthContextKey, authCtx)

		m.logger.Info().
			Str("user", user.Username).
			Str("userId", user.ID).
			Msg("authenticated user")

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth ensures a user is authenticated
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.enabled {
			next.ServeHTTP(w, r)
			return
		}

		authCtx, ok := GetAuthContextFromContext(r.Context())
		if !ok || authCtx.IsAnonymous {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// AuthorizeAction checks if the user can perform an action on a resource
func (m *Middleware) AuthorizeAction(resource string, action models.Action) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !m.enabled {
				next.ServeHTTP(w, r)
				return
			}

			authCtx, ok := GetAuthContextFromContext(r.Context())
			if !ok {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Authorize the action
			authorized, err := m.policyEngine.Authorize(authCtx.User, resource, action)
			if err != nil || !authorized {
				m.logger.Warn().
					Str("user", getUsername(authCtx.User)).
					Str("resource", resource).
					Str("action", string(action)).
					Err(err).
					Msg("authorization denied")

				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}

			m.logger.Debug().
				Str("user", getUsername(authCtx.User)).
				Str("resource", resource).
				Str("action", string(action)).
				Msg("authorization granted")

			next.ServeHTTP(w, r)
		})
	}
}

// ExtractResourceFromRequest extracts the resource identifier from the request
func ExtractResourceFromRequest(r *http.Request) string {
	path := strings.TrimPrefix(r.URL.Path, "/s3/")
	if path == "" {
		return "*"
	}
	return path
}

// MapMethodToAction maps HTTP methods to actions
func MapMethodToAction(method string) models.Action {
	switch method {
	case http.MethodGet, http.MethodHead:
		return models.ActionRead
	case http.MethodPut, http.MethodPost:
		return models.ActionWrite
	case http.MethodDelete:
		return models.ActionDelete
	default:
		return models.ActionRead
	}
}

// getUsername safely gets username from user
func getUsername(user *models.User) string {
	if user == nil {
		return "anonymous"
	}
	return user.Username
}
