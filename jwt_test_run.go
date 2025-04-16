package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// Default secret key for JWT tokens
	jwtSecretKey = []byte("test-jwt-secret-key")
)

// JWT claim structure
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
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
			Issuer:    "termpos-test",
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

func main() {
	// Generate a test token
	tokenString, err := GenerateJWT(1, "admin", "admin")
	if err != nil {
		fmt.Printf("Error generating token: %v\n", err)
		return
	}

	fmt.Printf("Generated token: %s\n\n", tokenString)

	// Validate the token
	claims, err := ValidateJWT(tokenString)
	if err != nil {
		fmt.Printf("Error validating token: %v\n", err)
		return
	}

	// Print the claims
	fmt.Println("Token validation successful!")
	fmt.Printf("User ID: %d\n", claims.UserID)
	fmt.Printf("Username: %s\n", claims.Username)
	fmt.Printf("Role: %s\n", claims.Role)
	fmt.Printf("Expires: %v\n", claims.ExpiresAt)
}