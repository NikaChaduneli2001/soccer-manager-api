package middleware

import (
	"net/http"
	"strings"

	"github.com/nika/soccer-manager-api/pkg/auth"
	"github.com/nika/soccer-manager-api/pkg/response"
)

// JWT returns a middleware that validates the Bearer token and sets user ID in context.
func JWT(secret string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			if authHeader == "" {
				response.Error(w, http.StatusUnauthorized, "missing authorization header")
				return
			}
			const prefix = "Bearer "
			tokenString := authHeader
			if strings.HasPrefix(authHeader, prefix) {
				tokenString = strings.TrimSpace(authHeader[len(prefix):])
			}
			if tokenString == "" {
				response.Error(w, http.StatusUnauthorized, "invalid authorization format")
				return
			}
			userID, err := auth.ParseToken(tokenString, secret)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}
			ctx := auth.WithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
