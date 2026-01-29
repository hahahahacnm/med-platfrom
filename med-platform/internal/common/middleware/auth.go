package middleware

import (
	"med-platform/internal/common/db"
	"med-platform/internal/common/jwt"
	"med-platform/internal/user" // 引入 user 包以访问 User 模型
	"net/http"
	"strings"
	"time" // 👈 必须引入 time 包用于比对封禁时间

	"github.com/gin-gonic/gin"
)

// AuthJWT 基础鉴权中间件：验证 Token + 检查封号状态
func AuthJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取 Authorization Header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "需要认证 Token"})
			c.Abort()
			return
		}

		// 2. 提取 Token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 格式错误"})
			c.Abort()
			return
		}

		// 3. 解析 Token
		claims, err := jwt.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 无效或已过期"})
			c.Abort()
			return
		}

		// 4. 获取 UserID
		var uid uint
		if val, ok := claims["user_id"].(float64); ok {
			uid = uint(val)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 数据异常"})
			c.Abort()
			return
		}

		// 🔥🔥🔥 5. 核心升级：查库校验用户状态 (封号拦截) 🔥🔥🔥
		var currentUser user.User
		if err := db.DB.First(&currentUser, uid).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
			return
		}

		// 6. 检查是否被封禁 (Status = 2)
		if currentUser.Status == 2 {
			// 如果有封禁截止时间，且当前时间还没到解封时间
			if currentUser.BanUntil != nil && time.Now().Before(*currentUser.BanUntil) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error":  "账号已被封禁",
					"reason": "违反平台规定",
					"until":  currentUser.BanUntil.Format("2006-01-02 15:04:05"), // 告诉用户什么时候解封
				})
				return
			}
			// 如果时间已过，原则上可以放行（或者你可以在这里写逻辑自动把 status 改回 1）
		}

		// 7. 将关键信息存入上下文，供后续使用
		c.Set("userID", currentUser.ID)
		c.Set("role", currentUser.Role) // 🔥 把角色存进去，AdminRequired 直接用，不用再查库了
		c.Set("username", currentUser.Username)

		c.Next()
	}
}

// 🔥🔥🔥 管理员权限验证中间件 (优化版) 🔥🔥🔥
// 必须在 AuthJWT 之后使用
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 直接从上下文获取 Role (AuthJWT 已经查过库了，这里直接用，性能更高)
		role, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}

		roleStr := role.(string)

		// 2. 权限判断
		// 允许 'admin'(超管) 和 'agent'(机构代理) 进入后台
		// 如果你只想让 admin 进，就去掉 agent
		if roleStr != "admin" && roleStr != "agent" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "权限不足：非管理员账号"})
			return
		}

		c.Next()
	}
}