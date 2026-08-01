package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yeung/system-hardening/backend/configs"
)

// Claims 定义 JWT 声明结构
type Claims struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	UserID   string `json:"user_id,omitempty"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT token
func GenerateToken(username, email, userID string, jwtConfig configs.JWTConfig) (string, error) {
	// 计算过期时间
	expirationTime := time.Now().Add(time.Duration(jwtConfig.ExpiryHour) * time.Hour)

	claims := &Claims{
		Username: username,
		Email:    email,
		UserID:   userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "system-hardening-platform",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	tokenString, err := token.SignedString([]byte(jwtConfig.SecretKey))
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return tokenString, nil
}

// ValidateToken 验证并解析 JWT token
func ValidateToken(tokenString string, jwtConfig configs.JWTConfig) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtConfig.SecretKey), nil
	})

	if err != nil {
		return nil, errors.New("invalid or expired token")
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}

// RefreshToken 刷新 token（可选功能）
func RefreshToken(oldToken string, jwtConfig configs.JWTConfig) (string, error) {
	claims, err := ValidateToken(oldToken, jwtConfig)
	if err != nil {
		return "", err
	}

	// 生成新 token
	newToken, err := GenerateToken(
		claims.Username,
		claims.Email,
		claims.UserID,
		jwtConfig,
	)
	if err != nil {
		return "", errors.New("failed to refresh token")
	}

	return newToken, nil
}
