package jwt

import (
	"errors"
	"time"

	"med-platform/internal/common/config"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateToken 生成 JWT Token
func GenerateToken(userID uint, username string) (string, error) {
	// 🔥 修改点：JWT -> Jwt (根据报错提示调整大小写)
	secret := []byte(config.GlobalConfig.Jwt.Secret)
	if len(secret) == 0 {
		secret = []byte("default_secret_key")
	}

	// 创建 Claims
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24 * 7).Unix(), // 7天过期
		"iat":      time.Now().Unix(),
	}

	// 生成 Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ParseToken 解析 JWT Token
func ParseToken(tokenString string) (jwt.MapClaims, error) {
	// 🔥 修改点：JWT -> Jwt
	secret := []byte(config.GlobalConfig.Jwt.Secret)
	if len(secret) == 0 {
		secret = []byte("default_secret_key")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}