package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// CorsMiddleware CORS中间件
type CorsMiddleware struct{}

// NewCorsMiddleware 创建CORS中间件
func NewCorsMiddleware() *CorsMiddleware {
	return &CorsMiddleware{}
}

// Handle CORS处理
func (c *CorsMiddleware) Handle() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		origin := ctx.Request.Header.Get("Origin")
		if origin != "" {
			ctx.Header("Access-Control-Allow-Origin", origin)
		} else {
			ctx.Header("Access-Control-Allow-Origin", "*")
		}

		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ctx.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		ctx.Header("Access-Control-Allow-Credentials", "true")

		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	secret string
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(secret string) *AuthMiddleware {
	return &AuthMiddleware{secret: secret}
}

// Handle 认证处理
func (a *AuthMiddleware) Handle() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 跳过认证的路径
		skipPaths := []string{
			"/api/v1/users/register",
			"/api/v1/users/login",
			"/api/v1/users/verify-email",
			"/api/v1/users/resend-code",
		}

		path := ctx.Request.URL.Path
		for _, skipPath := range skipPaths {
			if strings.HasPrefix(path, skipPath) {
				ctx.Next()
				return
			}
		}

		// JWT 验证
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid token"})
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return []byte(a.secret), nil
		})
		if err != nil || !token.Valid {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		ctx.Next()
	}
}
