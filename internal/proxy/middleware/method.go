package middleware

import (
	"net/http"
	"slices"
	"strings"
)

// NewRejectMethodMiddleware returns a middleware to reject requests with methods not in the allowedMethods list.
// If allowedMethods is empty, all methods are allowed.
func NewRejectMethodMiddleware(allowedMethods ...string) func(http.Handler) http.Handler {
	normalized := make([]string, 0, len(allowedMethods))
	for _, method := range allowedMethods {
		m := strings.ToUpper(strings.TrimSpace(method))
		if m == "" {
			continue
		}
		if !slices.Contains(normalized, m) {
			normalized = append(normalized, m)
		}
	}
	allowValue := strings.Join(normalized, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(normalized) == 0 {
				next.ServeHTTP(w, r)
				return
			}
			if slices.Contains(normalized, r.Method) {
				next.ServeHTTP(w, r)
				return
			}
			w.Header().Set("Allow", allowValue)
			w.WriteHeader(http.StatusMethodNotAllowed)
		})
	}
}
