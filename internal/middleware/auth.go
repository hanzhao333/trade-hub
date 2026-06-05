package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zhanghanzhao/trade-hub/internal/service"
	"github.com/zhanghanzhao/trade-hub/pkg/response"
)

const UserIDKey = "userID"

//	type AuthService struct {
//		users     *repository.UserRepo
//		jwtSecret []byte
//	}
//
//	type UserRepo struct {
//		db *gorm.DB
//	}
func Auth(auth *service.AuthService) gin.HandlerFunc {
	// 传入的c 也是*service.AuthService
	// 这是我从调用链路上得出来，但是我看这里代码又可以调用GetHeader方法
	// 然后还是*gin.Context类型，这是为啥
	return func(c *gin.Context) {
		// 通过.Use注册的方法调用的时候都会自动注入*gin.Context对象
		// main.go规则内的请求路径，都会调用这个方法
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Unauthorized(c, "missing bearer token")
			c.Abort()
			return
		}
		// 去掉Bearer前缀，拿到token
		token := strings.TrimPrefix(header, "Bearer ")
		// 解析拿到token和err
		userID, err := auth.ParseToken(token)
		if err != nil {
			response.Unauthorized(c, "invalid token")
			c.Abort()
			return
		}
		// c:这一次 HTTP 请求的上下文,将userID存储到c的上下文中,本次请求结束c即销毁
		// 假设前端访问/profile，这样算一次请求，等请求给前端返回数据即算结束
		c.Set(UserIDKey, userID)
		c.Next()
	}
}
