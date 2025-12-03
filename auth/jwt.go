package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var jwtSecret = []byte("epocheye")

const (
	defaultAccessTTL  = 30 * 24 * time.Hour
	defaultRefreshTTL = 7 * 24 * time.Hour
)

func GenerateJWT(userUUID uuid.UUID, email string, accessTTL, refreshTTL time.Duration) (uuid.UUID, string, string, time.Time, time.Time, error) {
	genAt := time.Now()
	accessExp := genAt.Add(defaultAccessTTL)
	refreshExp := genAt.Add(defaultRefreshTTL)

	accessClaims := jwt.MapClaims{
		"user_uuid": userUUID,
		"email":     email,
		"exp":       accessExp.Unix(),
		"type":      "access",
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(jwtSecret)
	if err != nil {
		return userUUID, "", "", genAt, accessExp, err
	}

	refreshClaims := jwt.MapClaims{
		"user_uuid": userUUID,
		"email":     email,
		"exp":       refreshExp.Unix(),
		"type":      "refresh",
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(jwtSecret)
	if err != nil {
		return userUUID, accessToken, "", genAt, accessExp, err
	}

	return userUUID, accessToken, refreshToken, genAt, accessExp, nil
}
func ValidateJWT(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		// ensure it's signed with HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrInvalidType
	}

	// Check expiration time
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, jwt.ErrTokenExpired
		}
	}

	return claims, nil
}
