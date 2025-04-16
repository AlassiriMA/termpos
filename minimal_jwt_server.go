package main

import (
        "encoding/json"
        "fmt"
        "net/http"
        "os"
        "strings"
        "time"

        "github.com/golang-jwt/jwt/v5"
)

var (
        // Default secret key for JWT tokens
        jwtSecretKey = []byte("test-jwt-secret-key")
        // Hardcoded user info for testing
        testUser = User{
                ID:       1,
                Username: "admin",
                Password: "password123",
                Role:     "admin",
        }
)

// User represents a user in the system
type User struct {
        ID       int    `json:"id"`
        Username string `json:"username"`
        Password string `json:"password"`
        Role     string `json:"role"`
}

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
                        Issuer:    "termpos-minimal",
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

// authMiddleware checks if the request has a valid authentication token
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
                // Get token from Authorization header
                authHeader := r.Header.Get("Authorization")
                if authHeader == "" {
                        http.Error(w, "Authorization header required", http.StatusUnauthorized)
                        return
                }

                // Expected format: "Bearer <token>"
                parts := strings.Split(authHeader, " ")
                if len(parts) != 2 || parts[0] != "Bearer" {
                        http.Error(w, "Invalid authorization format, expected 'Bearer <token>'", http.StatusUnauthorized)
                        return
                }

                // Extract the token
                tokenString := parts[1]

                // Validate the JWT token
                _, err := ValidateJWT(tokenString)
                if err != nil {
                        http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
                        return
                }

                // Call the next handler with the user information
                next(w, r)
        }
}

// Handle login requests
func handleLogin(w http.ResponseWriter, r *http.Request) {
        // Only accept POST method
        if r.Method != http.MethodPost {
                http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
                return
        }

        // Parse credentials from request body
        var creds struct {
                Username string `json:"username"`
                Password string `json:"password"`
        }

        if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
                http.Error(w, "Invalid request body", http.StatusBadRequest)
                return
        }

        // Authenticate user (Using hardcoded values for testing)
        if creds.Username != testUser.Username || creds.Password != testUser.Password {
                http.Error(w, "Invalid username or password", http.StatusUnauthorized)
                return
        }

        // Generate JWT token
        token, err := GenerateJWT(testUser.ID, testUser.Username, testUser.Role)
        if err != nil {
                http.Error(w, fmt.Sprintf("Failed to generate token: %v", err), http.StatusInternalServerError)
                return
        }

        // Return token and user information
        response := map[string]interface{}{
                "token": token,
                "user": map[string]interface{}{
                        "id":       testUser.ID,
                        "username": testUser.Username,
                        "role":     testUser.Role,
                },
                "expires_in": 86400, // 24 hours in seconds
        }

        // Set response headers and send response
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
}

// Protected endpoint returning a simple message
func handleProtected(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{
                "message": "This is a protected endpoint!",
                "time":    time.Now().Format(time.RFC3339),
        })
}

// Health check endpoint
func handleHealth(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{
                "status": "ok",
                "time":   time.Now().Format(time.RFC3339),
        })
}

func main() {
        port := 5000

        // Get port from environment variable if available
        if portEnv := os.Getenv("PORT"); portEnv != "" {
                fmt.Sscanf(portEnv, "%d", &port)
        }

        // Set up routes
        http.HandleFunc("/health", handleHealth)
        http.HandleFunc("/auth/login", handleLogin)
        http.HandleFunc("/protected", authMiddleware(handleProtected))

        // Start the server
        addr := fmt.Sprintf("0.0.0.0:%d", port)
        fmt.Printf("Starting minimal JWT server on %s\n", addr)
        if err := http.ListenAndServe(addr, nil); err != nil {
                fmt.Printf("Error starting server: %v\n", err)
                os.Exit(1)
        }
}