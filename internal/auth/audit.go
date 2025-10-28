package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/candlekeep/zot-artifact-store/internal/models"
	"github.com/candlekeep/zot-artifact-store/internal/storage"
	"github.com/google/uuid"
	"zotregistry.io/zot/pkg/log"
)

// AuditLogger logs all access attempts and API calls
type AuditLogger struct {
	store   *storage.MetadataStore
	logger  log.Logger
	enabled bool
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(store *storage.MetadataStore, logger log.Logger, enabled bool) *AuditLogger {
	return &AuditLogger{
		store:   store,
		logger:  logger,
		enabled: enabled,
	}
}

// LogAccess logs an access attempt
func (a *AuditLogger) LogAccess(r *http.Request, status int, user *models.User, err error) {
	if !a.enabled {
		return
	}

	log := &models.AuditLog{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Action:    r.Method,
		Resource:  r.URL.Path,
		Method:    r.Method,
		Status:    status,
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
		Metadata:  make(map[string]string),
	}

	if user != nil {
		log.UserID = user.ID
		log.Username = user.Username
	} else {
		log.UserID = "anonymous"
		log.Username = "anonymous"
	}

	if err != nil {
		log.Error = err.Error()
	}

	// Add query parameters to metadata
	if len(r.URL.Query()) > 0 {
		queryJSON, _ := json.Marshal(r.URL.Query())
		log.Metadata["query"] = string(queryJSON)
	}

	// Store audit log
	if err := a.store.StoreAuditLog(log); err != nil {
		a.logger.Error().Err(err).Msg("failed to store audit log")
	}

	// Also log to structured logger
	logEvent := a.logger.Info().
		Str("auditId", log.ID).
		Str("user", log.Username).
		Str("userId", log.UserID).
		Str("method", log.Method).
		Str("resource", log.Resource).
		Int("status", log.Status).
		Str("ip", log.IPAddress)

	if log.Error != "" {
		logEvent = logEvent.Str("error", log.Error)
	}

	logEvent.Msg("audit log")
}

// LogAccessWithUser is a convenience method that logs access with user from context
func (a *AuditLogger) LogAccessFromContext(r *http.Request, status int, err error) {
	user, _ := GetUserFromContext(r.Context())
	a.LogAccess(r, status, user, err)
}

// GetAuditLogs retrieves audit logs with optional filtering
func (a *AuditLogger) GetAuditLogs(userID string, resource string, startTime, endTime time.Time, limit int) ([]*models.AuditLog, error) {
	return a.store.ListAuditLogs(userID, resource, startTime, endTime, limit)
}

// AuditMiddleware wraps an HTTP handler with audit logging
func (a *AuditLogger) AuditMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wrap response writer to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Get user from context before processing request
		user, _ := GetUserFromContext(r.Context())

		// Process request
		next.ServeHTTP(wrapped, r)

		// Log the access
		a.LogAccess(r, wrapped.statusCode, user, nil)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to remote address
	return r.RemoteAddr
}
