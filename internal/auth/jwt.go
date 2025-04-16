package auth

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// Default secret key for JWT tokens - this should be overridden in production
	jwtSecretKey = []byte("termpos-default-secret-key-change-in-production")
)

// JWT claim structure
type Claims struct {
	UserID   int        `json:"user_id"`
	Username string     `json:"username"`
	Role     string     `json:"role"`
	jwt.RegisteredClaims
}

// Initialize JWT configuration
func InitJWT() {
	// Check if a custom secret is provided via environment variable
	if secretKey := os.Getenv("TERMPOS_JWT_SECRET"); secretKey != "" {
		jwtSecretKey = []byte(secretKey)
		fmt.Println("Using JWT secret from environment variable")
	} else {
		fmt.Println("Warning: Using default JWT secret key. Please set TERMPOS_JWT_SECRET environment variable.")
	}
}

// GenerateJWT creates a new JWT token for a user
func GenerateJWT(userID int, username string, role string) (string, error) {
	// Token expiration time (24 hours)
	expirationTime := time.Now().Add(24 * time.Hour)

	// Create claims with user data
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "termpos",
			Subject:   username,
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with our secret key
	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT validates a JWT token and returns the claims
func ValidateJWT(tokenString string) (*Claims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecretKey, nil
		},
	)

	if err != nil {
		return nil, err
	}

	// Check if the token is valid
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract the claims
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}