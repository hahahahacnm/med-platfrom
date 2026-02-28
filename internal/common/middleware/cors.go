package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware 处理跨域请求
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 允许的源 (生产环境建议改成具体的域名，开发环境用 *)
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		
		// 允许的方法
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		
		// 允许的头部 (这一步很关键！前端传了 Authorization，这里必须允许)
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		// 🔥🔥🔥 核心解决代码 🔥🔥🔥
		// 如果是 OPTIONS 请求，直接返回 204 No Content，并终止后续处理
		// 这样 Gin 就不会去路由表里找 OPTIONS 的 handler 了，也就不会报 404 了
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}