package middleware

import (
	"chat_backend/internal/global"
	"chat_backend/internal/response"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// JWTClaims 自定义JWT声明
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// JWTConfig JWT配置
type JWTConfig struct {
	SecretKey     string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

var jwtConfig *JWTConfig

// GetJWTConfig 获取JWT配置（用于其他包访问）
func GetJWTConfig() *JWTConfig {
	return jwtConfig
}

// InitJWT 初始化JWT配置
func InitJWT(secretKey string, accessExpiry, refreshExpiry time.Duration) {
	jwtConfig = &JWTConfig{
		SecretKey:     secretKey,
		AccessExpiry:  accessExpiry,
		RefreshExpiry: refreshExpiry,
	}
}

// GenerateAccessToken 生成访问令牌
func GenerateAccessToken(userID, username string) (string, error) {
	if jwtConfig == nil {
		return "", echo.NewHTTPError(http.StatusInternalServerError, "JWT not initialized")
	}

	claims := &JWTClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtConfig.AccessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtConfig.SecretKey))
}

// GenerateRefreshToken 生成刷新令牌
func GenerateRefreshToken(userID string) (string, error) {
	if jwtConfig == nil {
		return "", echo.NewHTTPError(http.StatusInternalServerError, "JWT not initialized")
	}

	claims := &jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtConfig.RefreshExpiry)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtConfig.SecretKey))
}

// ValidateToken 验证令牌
func ValidateToken(tokenString string) (*JWTClaims, error) {
	if jwtConfig == nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "JWT not initialized")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, echo.NewHTTPError(http.StatusUnauthorized, "unexpected signing method")
		}
		return []byte(jwtConfig.SecretKey), nil
	})

	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
}

// JWTMiddleware JWT认证中间件
func JWTMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 优先从Header获取，如果没有则从URL参数获取
			authHeader := c.Request().Header.Get("Authorization")
			var tokenString string

			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == global.JwtBearerPrefix {
					tokenString = parts[1]
				}
			} else {
				// 从URL参数获取
				tokenString = c.QueryParam("token")
				if tokenString == "" {
					return response.Error(c, 2000, "missing authorization header or token parameter")
				}
			}
			claims, err := ValidateToken(tokenString)
			if err != nil {
				return response.Error(c, 2001, "invalid or expired token")
			}

			// 将用户信息存储到上下文中
			c.Set(global.JwtKeyUserID, claims.UserID)
			c.Set(global.JwtKeyUserName, claims.Username)
			c.Set(global.JwtKeyClaims, claims)

			return next(c)
		}
	}
}
