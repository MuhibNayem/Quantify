package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

var jwtKey []byte
var refreshTokenSecret []byte

func InitializeJWT(secret, refreshSecret string) {
	jwtKey = []byte(secret)
	refreshTokenSecret = []byte(refreshSecret)
}

type Claims struct {
	UserID uint   `json:"userId"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateTokens(userID uint, role string) (string, string, string, error) {
	accessToken, accessJTI, err := generateAccessToken(userID, role)
	if err != nil {
		return "", "", "", err
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return "", "", "", err
	}

	return accessToken, refreshToken, accessJTI, nil
}

func generateAccessToken(userID uint, role string) (string, string, error) {
	expirationTime := time.Now().Add(8 * time.Hour)
	jti := uuid.NewString()
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	return tokenString, jti, err
}

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func ValidateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// GetTokenClaims parses the token and returns the claims without validating the signature or expiration.
// This is useful for extracting information like JTI for logout.
func GetTokenClaims(tokenString string) (*Claims, error) {
	claims := &Claims{}
	_, _, err := new(jwt.Parser).ParseUnverified(tokenString, claims)
	if err != nil {
		return nil, err
	}
	return claims, nil
}
