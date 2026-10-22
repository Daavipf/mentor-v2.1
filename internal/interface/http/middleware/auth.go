package middleware

import (
	"context"
	"net/http"
	"strings"

	"mentor-api/internal/domain"
	"mentor-api/internal/infrastructure/security"
	"mentor-api/internal/interface/http/response"
)

type contextKey string

const (
	UserIDKey    contextKey = "userID"
	UserEmailKey contextKey = "userEmail"
)

func Auth(tokenManager security.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.SendError(w, domain.ErrUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				response.SendError(w, domain.ErrUnauthorized)
				return
			}

			tokenStr := parts[1]
			claims, err := tokenManager.ValidateToken(tokenStr)
			if err != nil {
				response.SendError(w, domain.ErrUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(UserIDKey).(string)
	return val, ok
}

func GetUserEmail(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(UserEmailKey).(string)
	return val, ok
}

