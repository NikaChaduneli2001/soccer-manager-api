package middleware

import (
	"net/http"
	"strings"

	"github.com/nika/soccer-manager-api/pkg/response"
)

type Middleware func(http.Handler) http.Handler

// Chain returns a single middleware that applies all of the given middlewares.
func Chain(ms ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(ms) - 1; i >= 0; i-- {
			next = ms[i](next)
		}
		return next
	}
}

func Method(methods ...string) Middleware {
	allowed := make(map[string]bool)
	for _, m := range methods {
		allowed[strings.ToUpper(m)] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !allowed[r.Method] {
				response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
