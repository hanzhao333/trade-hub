package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zhanghanzhao/trade-hub/internal/service"
	"github.com/zhanghanzhao/trade-hub/pkg/response"
)

const UserIDKey = "userID"

func Auth(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Unauthorized(c, "missing bearer token")
			c.Abort()
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		userID, err := auth.ParseToken(token)
		if err != nil {
			response.Unauthorized(c, "invalid token")
			c.Abort()
			return
		}
		c.Set(UserIDKey, userID)
		c.Next()
	}
}
