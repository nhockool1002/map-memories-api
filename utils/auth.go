package utils

import (
	"errors"
	"strings"
	"time"

	"map-memories-api/config"
	"map-memories-api/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type JWTClaims struct {
	UserID   uint      `json:"user_id"`
	UserUUID uuid.UUID `json:"user_uuid"`
	Email    string    `json:"email"`
	Username string    `json:"username"`
	jwt.RegisteredClaims
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// VerifyPassword verifies a password against its hash
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// GenerateJWT generates a JWT token for a user
func GenerateJWT(user *models.User) (string, error) {
	claims := JWTClaims{
		UserID:   user.ID,
		UserUUID: user.UUID,
		Email:    user.Email,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.AppConfig.JWT.Expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "map-memories-api",
			Subject:   user.UUID.String(),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWT.Secret))
}

// VerifyJWT verifies and parses a JWT token
func VerifyJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(config.AppConfig.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ExtractBearerToken extracts the token from the Authorization header
func ExtractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	// Handle both "Bearer <token>" and just "<token>" formats
	const bearerPrefix = "Bearer "
	if len(authHeader) >= len(bearerPrefix) && authHeader[:len(bearerPrefix)] == bearerPrefix {
		// Standard Bearer token format
		return authHeader[len(bearerPrefix):], nil
	}

	// If no Bearer prefix, assume the entire header is the token
	// This handles cases where Swagger UI doesn't add the prefix
	if strings.TrimSpace(authHeader) != "" {
		return strings.TrimSpace(authHeader), nil
	}

	return "", errors.New("invalid authorization header format")
}

// GenerateRefreshToken generates a refresh token (simple UUID for now)
func GenerateRefreshToken() string {
	return uuid.New().String()
}

// ValidateRefreshToken validates a refresh token format
func ValidateRefreshToken(token string) error {
	_, err := uuid.Parse(token)
	if err != nil {
		return errors.New("invalid refresh token format")
	}
	return nil
}