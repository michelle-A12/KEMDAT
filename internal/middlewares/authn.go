package middlewares

import (
	"aaa_bekend/internal/auth"
	"aaa_bekend/internal/lmodel"
	"net/http"
	"strings"
)

func AuthN(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		// TODO
		if authHeader == "" {
			http.Error(w, "Unauthorized: missing Authorization header", 			http.StatusUnauthorized)
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || 			strings.TrimSpace(parts[1]) == "" {
			http.Error(w, "Unauthorized: invalid Authorization header 			format", http.StatusUnauthorized)
			return
		}
		claims, err := auth.ValidateToken(strings.TrimSpace(parts[1]))
		if err != nil {
			http.Error(w, "Unauthorized: invalid or expired token", 		http.StatusUnauthorized)
			return
		}

		// TODO
		userCtx := lmodel.UserContext{
			ID:       claims.UserID,
			Username: claims.Username,
			Role:     claims.Role,
		}

		ctx := lmodel.ContextWithUser(r.Context(), userCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}