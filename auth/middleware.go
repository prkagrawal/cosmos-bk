// auth/middleware.go
package auth

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Define a custom key type for context to avoid collisions
type contextKey string

const userContextKey = contextKey("userContextKey")

// AuthMiddleware decodes the share session and packs the session into context
func AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			// Check if Authorization header is provided
			if authHeader == "" {
				next.ServeHTTP(w, r) // No token, proceed without user in context
				return
			}

			// Check if the token is a Bearer token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				// Optionally, you could return an error here (e.g., http.StatusUnauthorized)
				// For GraphQL, often it's better to let requests proceed and let resolvers handle auth.
				// If it's malformed, it's probably better to just ignore it or log it.
				next.ServeHTTP(w, r) // Malformed header, proceed without user
				return
			}

			tokenStr := parts[1]

			// Parse and validate the token
			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				// Make sure the signing method is what you expect:
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid // Or a more specific error
				}
				return []byte(os.Getenv("JWT_SECRET")), nil
			})

			if err != nil || !token.Valid {
				// Token is invalid or expired.
				// You could send an HTTP 401 here, but for GraphQL,
				// it's often preferred to let the request proceed.
				// Resolvers requiring auth will then fail if no user is in context.
				// log.Printf("Invalid or expired token: %v", err) // For debugging
				next.ServeHTTP(w, r) // Invalid token, proceed without user
				return
			}

			// Token is valid, add it to the context
			// We store the whole token object, as GetUserFromContext expects it.
			// Alternatively, you could extract claims (like user ID) and store that.
			ctxWithUser := context.WithValue(r.Context(), userContextKey, token)
			rWithUser := r.WithContext(ctxWithUser)

			next.ServeHTTP(w, rWithUser)
		})
	}
}
