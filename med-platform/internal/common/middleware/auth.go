package middleware

import (
	"med-platform/internal/common/db"
	"med-platform/internal/common/jwt"
	"med-platform/internal/user" 
	"net/http"
	"strings"
	"time" 

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

		// 5. 查库校验用户状态 (封号拦截)
		var currentUser user.User
		if err := db.DB.First(&currentUser, uid).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
			return
		}

		// 6. 检查是否被封禁 (Status = 2)
		if currentUser.Status == 2 {
			if currentUser.BanUntil != nil && time.Now().Before(*currentUser.BanUntil) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error":  "账号已被封禁",
					"reason": "违反平台规定",
					"until":  currentUser.BanUntil.Format("2006-01-02 15:04:05"),
				})
				return
			}
		}

		// 7. 将关键信息存入上下文
		c.Set("userID", currentUser.ID)
		c.Set("role", currentUser.Role) 
		c.Set("username", currentUser.Username)

		c.Next()
	}
}

// ---------------------------------------------------------
// 🔥 权限中间件升级区
// ---------------------------------------------------------

// RequireSuperAdmin 严格模式：仅限超级管理员 (Role = "admin")
// 用于：封号、修改题库、删库跑路级别操作
func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "权限不足：仅限超级管理员操作"})
			return
		}
		c.Next()
	}
}

// RequireStaff 宽松模式：内部工作人员 (Role = "admin" OR "agent")
// 用于：商品授权、查看日志、社区删帖、查看反馈
func RequireStaff() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		
		roleStr := role.(string)
		// 允许 admin 和 agent
		if roleStr != "admin" && roleStr != "agent" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "权限不足：仅限内部人员操作"})
			return
		}
		c.Next()
	}
}

// 兼容旧代码的别名 (如果不想改 router 里的名字，可以留着这个，指向 RequireStaff)
// 但建议我们在下一步直接替换 router 里的调用
func AdminRequired() gin.HandlerFunc {
    return RequireStaff() 
}