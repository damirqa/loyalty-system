package http

import (
	"context"
	"net/http"
	"strconv"
)

type contextKey string

const userIDKey = contextKey("userID")

func AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userIDHeader := r.Header.Get("X-User-ID")
			if userIDHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			userID, err := strconv.ParseInt(userIDHeader, 10, 64)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
