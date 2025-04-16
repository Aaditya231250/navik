package handlers

import (
    "context"
    "net/http"
    "strings"
    
    "authentication/internal/service"
)

func AuthMiddleware(authService *service.AuthService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract token from Authorization header
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
                respondWithError(w, http.StatusUnauthorized, "Authorization header is required")
                return
            }
            
            token := strings.TrimPrefix(authHeader, "Bearer ")
            
            // Validate the token
            claims, err := authService.ValidateAccessToken(token)
            if err != nil {
                if err == service.ErrExpiredToken {
                    respondWithError(w, http.StatusUnauthorized, "Token has expired")
                } else {
                    respondWithError(w, http.StatusUnauthorized, "Invalid token")
                }
                return
            }
            
            // Extract user ID and role from claims
            userID, _ := claims["user_id"].(string)
            userType, _ := claims["user_type"].(string)
            
            // Add user information to request context
            ctx := context.WithValue(r.Context(), "user_id", userID)
            ctx = context.WithValue(ctx, "user_type", userType)
            
            // Proceed with the next handler
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Get user type from context
            userType, ok := r.Context().Value("user_type").(string)
            if !ok {
                respondWithError(w, http.StatusUnauthorized, "Unauthorized")
                return
            }
            
            allowed := false
            for _, role := range allowedRoles {
                if userType == role {
                    allowed = true
                    break
                }
            }
            
            if !allowed {
                respondWithError(w, http.StatusForbidden, "Insufficient permissions")
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
