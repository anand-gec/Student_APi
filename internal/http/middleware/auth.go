package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/anand-gec/student-api/internal/config/utils/jwt"
	"github.com/anand-gec/student-api/internal/utils/response"
)

type contextKey string

const AdminContentKey contextKey = "admin_claims"

func RequireAdminAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// return 401 missing header
			return
		}

		// Strip "Bearer " prefix
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		tokenStr = strings.TrimSpace(tokenStr)

		// Pass only the raw token string
		claims, err := jwt.ValidateToken(tokenStr)
		if err != nil {
			response.WriteJson(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized:invalid token"})
			return
		}
		// Inject the authenticated admin ID into the request context
		ctx := context.WithValue(r.Context(), AdminContentKey, claims)
		next(w, r.WithContext(ctx))
	}
}
