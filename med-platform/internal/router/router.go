package router

import (
	"med-platform/internal/answer"
	"med-platform/internal/common/middleware"
	"med-platform/internal/note"
	"med-platform/internal/payment" // 🔥 引入支付模块
	"med-platform/internal/product"
	"med-platform/internal/question"
	"med-platform/internal/user"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 允许跨域
	r.Use(middleware.CORSMiddleware())
	// 静态文件服务
	r.Static("/uploads", "./uploads")

	// 初始化各模块 Handler
	userHandler := user.NewHandler()
	questionHandler := question.NewHandler()
	answerHandler := answer.NewHandler()
	
	// 🔥 修正 1：改回结构体初始化
	noteHandler := &note.Handler{}       
	
	productHandler := product.NewHandler()
	paymentHandler := payment.NewHandler() // 🔥 初始化支付 Handler

	api := r.Group("/api/v1")
	{
		// 🟢 公共区 (无需登录)
		api.POST("/auth/register", userHandler.Register)
		api.POST("/auth/login", userHandler.Login)
		api.GET("/category-tree", questionHandler.GetTree)

		// 💰 支付回调区
		api.GET("/payment/mock/callback", paymentHandler.MockSuccess)
		// api.POST("/payment/alipay/notify", paymentHandler.AlipayNotify)

		// 🟠 用户区 (需登录)
		userGroup := api.Group("/")
		userGroup.Use(middleware.AuthJWT())
		{
			// --- 👤 个人中心 ---
			userGroup.GET("/user/profile", userHandler.GetProfile)
			userGroup.PUT("/user/profile", userHandler.UpdateProfile)
			userGroup.POST("/user/avatar", userHandler.UploadAvatar)
			userGroup.PUT("/user/password", userHandler.ChangePassword)

			// --- 📝 题目与练习 ---
			userGroup.GET("/questions", questionHandler.List)
			userGroup.GET("/questions/:id", questionHandler.GetDetail)
			userGroup.GET("/banks", questionHandler.GetSources)
			userGroup.POST("/questions/:id/submit", answerHandler.Submit)

			// --- 📊 错题与统计 ---
			userGroup.GET("/mistakes", answerHandler.GetMistakes)
			userGroup.DELETE("/mistakes/:id", answerHandler.RemoveMistake)
			userGroup.GET("/mistake-tree", answerHandler.GetMistakeTree)
			userGroup.GET("/stats", answerHandler.GetStats)
			userGroup.GET("/rank/daily", answerHandler.GetDailyRank)

			// --- ⭐ 收藏 ---
			userGroup.POST("/favorites/:id", answerHandler.ToggleFavorite)
			userGroup.GET("/favorites", answerHandler.GetFavorites)
			userGroup.GET("/favorite-tree", answerHandler.GetFavoriteTree)

			// --- 🔄 重置 ---
			userGroup.DELETE("/questions/:id/reset", answerHandler.Reset)
			userGroup.DELETE("/answers/reset-chapter", answerHandler.ResetChapter)

			// --- 📓 笔记 ---
			userGroup.POST("/notes", noteHandler.SaveNote)       
			userGroup.GET("/notes", noteHandler.ListNotes)      
			userGroup.GET("/notes/my", noteHandler.GetMyNotes)   
			userGroup.GET("/notes/note-tree", noteHandler.GetNoteTree)
			userGroup.DELETE("/notes/:id", noteHandler.DeleteNote)
			userGroup.POST("/notes/:id/like", noteHandler.ToggleLike)
			userGroup.POST("/notes/:id/collect", noteHandler.ToggleCollect)

			// --- 💰 支付与商城 (Market) ---
			// 🔥🔥🔥 修改：使用 ListMarketProducts (只返回上架商品) 🔥🔥🔥
			// 旧代码: userGroup.GET("/market/products", productHandler.ListProducts)
			userGroup.GET("/market/products", productHandler.ListMarketProducts)
			
			userGroup.POST("/pay/create", paymentHandler.CreatePay)
			userGroup.GET("/user/products/:id", productHandler.GetUserProducts)

			// 🔴 管理员区 (需 Admin 权限)
			adminGroup := userGroup.Group("/admin")
			adminGroup.Use(middleware.AdminRequired())
			{
				// 👥 用户管理
				adminGroup.GET("/users", userHandler.ListUsers)
				adminGroup.POST("/users/role", userHandler.UpdateRole)
				adminGroup.POST("/users/ban", userHandler.BanUser)
				adminGroup.POST("/users/unban", userHandler.UnbanUser)

				// 🔥 上帝视角
				adminGroup.GET("/users/:id", userHandler.AdminGetUserDetail)
				adminGroup.PUT("/users/:id", userHandler.AdminUpdateUserInfo)
				adminGroup.PUT("/users/:id/password", userHandler.AdminResetPassword)
				adminGroup.POST("/users/:id/avatar", userHandler.AdminUploadAvatar)

				// 📦 商品系统
				adminGroup.POST("/products", productHandler.CreateProduct)
				adminGroup.GET("/products", productHandler.ListProducts)
				adminGroup.PUT("/products/:id", productHandler.UpdateProduct)
				adminGroup.DELETE("/products/:id", productHandler.DeleteProduct)
				adminGroup.GET("/auth-logs", productHandler.GetAuthLogs)

				// 🔗 内容绑定
				adminGroup.POST("/products/bind", productHandler.BindContent)
				adminGroup.POST("/products/unbind", productHandler.UnbindContent)
				adminGroup.GET("/products/:id/contents", productHandler.GetProductContents)

				// 🎫 授权管理
				adminGroup.POST("/users/grant", productHandler.GrantProductToUser)
				adminGroup.POST("/users/revoke", productHandler.RevokeUserProduct)
				adminGroup.GET("/users/:id/products", productHandler.GetUserProducts)

				// 📚 题库维护
				adminGroup.POST("/banks/rename", questionHandler.RenameSource)
				adminGroup.POST("/banks/delete", questionHandler.DeleteSource)
				adminGroup.POST("/banks/transfer", questionHandler.TransferCategory)
				adminGroup.POST("/categories/sync", questionHandler.SyncCategories)
				adminGroup.PUT("/categories/:id", questionHandler.UpdateCategory)
				adminGroup.POST("/categories/reorder", questionHandler.ReorderCategories)
				adminGroup.POST("/questions/import", questionHandler.ImportQuestions)
				adminGroup.PUT("/questions/:id", questionHandler.UpdateQuestion)
				adminGroup.DELETE("/questions/:id", questionHandler.DeleteQuestion)
				adminGroup.POST("/questions/batch-delete", questionHandler.BatchDeleteQuestions)
				adminGroup.DELETE("/questions/by-category", questionHandler.DeleteByCategory)
			}
		}
	}

	return r
}