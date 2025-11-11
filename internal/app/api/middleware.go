// Package api provides the HTTP API implementation for the VPN Automation System.
package api

import (
	"context"
	"net/http"
	"strings"

	"D:/vista-app/internal/app/auth"
)

// Context key for username.
type contextKey string

const contextKeyUsername contextKey = "username"

// authMiddleware validates JWT tokens from requests.
func (r *Router) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Extract token from Authorization header
		authHeader := req.Header.Get("Authorization")
		if authHeader == "" {
			r.logger.Warn("Missing Authorization header")
			respondError(w, http.StatusUnauthorized, "Missing authorization")
			return
		}

		// Format should be "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			r.logger.Warn("Malformed Authorization header", "header", authHeader)
			respondError(w, http.StatusUnauthorized, "Invalid authorization format")
			return
		}

		token := parts[1]
		claims, err := r.jwtSvc.ValidateToken(token)
		if err != nil {
			r.logger.Warn("Invalid token", "error", err)
			respondError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// Add username to request context
		ctx := context.WithValue(req.Context(), contextKeyUsername, claims.Username)
		r.logger.Debug("Authentication successful", "username", claims.Username)

		// Continue with authenticated request
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}