package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("epocheye")

const (
	defaultAccessTTL  = 2 * time.Hour
	defaultRefreshTTL = 7 * 24 * time.Hour
)

// GenerateJWT — creates access + refresh tokens
func GenerateJWT(email string, accessTTL, refreshTTL time.Duration) (string, string, time.Time, time.Time, error) {
	genAt := time.Now()
	accessExp := genAt.Add(defaultAccessTTL)
	refreshExp := genAt.Add(defaultRefreshTTL)

	// Access token
	accessClaims := jwt.MapClaims{
		"email": email,
		"exp":   accessExp.Unix(),
		"type":  "access",
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(jwtSecret)
	if err != nil {
		return "", "", genAt, accessExp, err
	}

	// Refresh token
	refreshClaims := jwt.MapClaims{
		"email": email,
		"exp":   refreshExp.Unix(),
		"type":  "refresh",
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(jwtSecret)
	if err != nil {
		return "", "", genAt, accessExp, err
	}

	return accessToken, refreshToken, genAt, accessExp, nil
}

// ValidateJWT validates a JWT token and returns the claims
func ValidateJWT(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Check if token is expired
	if exp, ok := claims["exp"]; ok {
		if expTime, ok := exp.(float64); ok {
			if time.Now().Unix() > int64(expTime) {
				return nil, errors.New("token expired")
			}
		}
	}

	return claims, nil
}
