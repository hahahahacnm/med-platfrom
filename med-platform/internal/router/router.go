package router

import (
	"med-platform/internal/answer"
	"med-platform/internal/common/middleware"
	"med-platform/internal/feedback"
	"med-platform/internal/forum"
	"med-platform/internal/note"
	"med-platform/internal/payment"
	"med-platform/internal/product"
	"med-platform/internal/question"
	"med-platform/internal/user"

	"github.com/gin-gonic/gin"
)

// SetupRouter 初始化路由
func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(middleware.CORSMiddleware())
	r.Static("/uploads", "./uploads")

	commentLimiter := middleware.NewIPRateLimiter(1, 3)
	uploadLimiter := middleware.NewIPRateLimiter(2, 5)

	userHandler := user.NewHandler()
	questionHandler := question.NewHandler()
	answerHandler := answer.NewHandler()
	noteHandler := &note.Handler{}
	productHandler := product.NewHandler()
	paymentHandler := payment.NewHandler()
	feedbackHandler := feedback.NewHandler()
	forumHandler := forum.NewHandler()

	api := r.Group("/api/v1")
	{
		// 🟢 公共区
		api.POST("/auth/register", userHandler.Register)
		api.POST("/auth/login", userHandler.Login)
		api.GET("/category-tree", questionHandler.GetTree)
		api.GET("/payment/mock/callback", paymentHandler.MockSuccess)

		// 🟠 用户区 (需登录)
		userGroup := api.Group("/")
		userGroup.Use(middleware.AuthJWT())
		{
			// 上传
			userGroup.POST("/upload/avatar", middleware.RateLimitMiddleware(uploadLimiter), userHandler.UploadAvatar)
			userGroup.POST("/upload/payment", middleware.RateLimitMiddleware(uploadLimiter), userHandler.UploadPaymentCode)
			userGroup.POST("/upload", middleware.RateLimitMiddleware(uploadLimiter), userHandler.UploadAvatar)

			// 消息/论坛
			userGroup.GET("/notifications", forumHandler.GetNotifications)
			userGroup.PUT("/notifications/:id/read", forumHandler.ReadNotification)
			userGroup.PUT("/notifications/read-all", forumHandler.ReadAllNotifications)

			userGroup.GET("/forum/boards", forumHandler.ListBoards)
			userGroup.GET("/forum/posts", forumHandler.ListPosts)
			userGroup.GET("/forum/posts/:id", forumHandler.GetPostDetail)
			userGroup.POST("/forum/posts", forumHandler.CreatePost)
			userGroup.DELETE("/forum/posts/:id", forumHandler.DeletePost)
			userGroup.POST("/forum/upload", middleware.RateLimitMiddleware(uploadLimiter), forumHandler.UploadImage)
			userGroup.GET("/forum/comments", forumHandler.ListComments)
			userGroup.POST("/forum/comments", forumHandler.CreateComment)
			userGroup.DELETE("/forum/comments/:id", forumHandler.DeleteComment)
			userGroup.POST("/forum/report", forumHandler.CreateReport)

			// 个人
			userGroup.GET("/user/profile", userHandler.GetProfile)
			userGroup.PUT("/user/profile", userHandler.UpdateProfile)
			userGroup.POST("/user/avatar", userHandler.UploadAvatar) 
			userGroup.PUT("/user/password", userHandler.ChangePassword)

			// 题目
			userGroup.GET("/questions", questionHandler.List)
			userGroup.GET("/questions/:id", questionHandler.GetDetail)
			userGroup.GET("/banks", questionHandler.GetSources)
			userGroup.POST("/questions/:id/submit", answerHandler.Submit)
			userGroup.POST("/feedback", questionHandler.SubmitFeedback)

			// 统计/收藏
			userGroup.GET("/mistakes", answerHandler.GetMistakes)
			userGroup.DELETE("/mistakes/:id", answerHandler.RemoveMistake)
			userGroup.GET("/mistake-tree", answerHandler.GetMistakeTree)
			userGroup.GET("/stats", answerHandler.GetStats)
			userGroup.GET("/rank/daily", answerHandler.GetDailyRank)
			userGroup.POST("/favorites/:id", answerHandler.ToggleFavorite)
			userGroup.GET("/favorites", answerHandler.GetFavorites)
			userGroup.GET("/favorite-tree", answerHandler.GetFavoriteTree)

			// 重置/笔记
			userGroup.DELETE("/questions/:id/reset", answerHandler.Reset)
			userGroup.DELETE("/answers/reset-chapter", answerHandler.ResetChapter)
			userGroup.POST("/notes", middleware.RateLimitMiddleware(commentLimiter), noteHandler.SaveNote)
			userGroup.GET("/notes", noteHandler.ListNotes)
			userGroup.GET("/notes/my", noteHandler.GetMyNotes)
			userGroup.GET("/notes/note-tree", noteHandler.GetNoteTree)
			userGroup.DELETE("/notes/:id", noteHandler.DeleteNote)
			userGroup.POST("/notes/upload", middleware.RateLimitMiddleware(uploadLimiter), noteHandler.UploadImage)
			userGroup.POST("/notes/:id/like", noteHandler.ToggleLike)
			userGroup.POST("/notes/:id/collect", noteHandler.ToggleCollect)
			userGroup.POST("/notes/:id/report", noteHandler.ReportNote)

			// 支付/反馈
			userGroup.GET("/market/products", productHandler.ListMarketProducts)
			userGroup.POST("/pay/create", paymentHandler.CreatePay)
			userGroup.GET("/user/products/:id", productHandler.GetUserProducts)
			userGroup.POST("/platform-feedback", feedbackHandler.Create)
			userGroup.GET("/platform-feedback", feedbackHandler.GetMyList)

			// ============================================================
			// 🔴 后台管理区
			// ============================================================
			
			// 1️⃣ 员工组
			staffGroup := userGroup.Group("/admin")
			staffGroup.Use(middleware.RequireStaff())
			{
				staffGroup.GET("/products", productHandler.ListProducts) 
				staffGroup.POST("/products/bind", productHandler.BindContent) 
				staffGroup.POST("/products/unbind", productHandler.UnbindContent)
				staffGroup.POST("/users/grant", productHandler.GrantProductToUser)
				staffGroup.POST("/users/revoke", productHandler.RevokeUserProduct)
				
				staffGroup.GET("/users", userHandler.ListUsers) 
				staffGroup.GET("/users/:id", userHandler.AdminGetUserDetail)
				staffGroup.GET("/users/:id/products", productHandler.GetUserProducts)
				staffGroup.GET("/auth-logs", productHandler.GetAuthLogs)

				staffGroup.GET("/dashboard/stats", userHandler.GetDashboardStats)
				staffGroup.POST("/withdraw/apply", userHandler.ApplyWithdraw)

				staffGroup.GET("/forum/comments", forumHandler.AdminListComments)
				staffGroup.DELETE("/forum/comments/:id", forumHandler.DeleteComment)
				staffGroup.DELETE("/forum/posts/:id", forumHandler.DeletePost)
				
				staffGroup.GET("/forum/reports", forumHandler.AdminListReports)
				staffGroup.GET("/forum/reports/preview", forumHandler.AdminGetReportContent)
				staffGroup.PUT("/forum/reports/:id/resolve", forumHandler.AdminResolveReport)
				
				staffGroup.GET("/notes", noteHandler.AdminListNotes)
				staffGroup.POST("/notes/:id/ignore", noteHandler.AdminDismissReport)

				staffGroup.GET("/feedbacks", questionHandler.AdminListFeedbacks)
				staffGroup.PUT("/feedbacks/:id", questionHandler.AdminResolveFeedback)
				staffGroup.GET("/platform-feedbacks", feedbackHandler.AdminList)
				staffGroup.PUT("/platform-feedbacks/:id", feedbackHandler.AdminReply)

				// 2️⃣ 超级管理员组
				superGroup := staffGroup.Group("/") 
				superGroup.Use(middleware.RequireSuperAdmin())
				{
					superGroup.POST("/users/role", userHandler.UpdateRole) 
					superGroup.POST("/users/ban", userHandler.BanUser)
					superGroup.POST("/users/unban", userHandler.UnbanUser)
					
					superGroup.PUT("/users/:id", userHandler.AdminUpdateUserInfo)
					superGroup.PUT("/users/:id/password", userHandler.AdminResetPassword)
					superGroup.POST("/users/:id/avatar", userHandler.AdminUploadAvatar)

					superGroup.POST("/products", productHandler.CreateProduct)
					superGroup.PUT("/products/:id", productHandler.UpdateProduct)
					superGroup.DELETE("/products/:id", productHandler.DeleteProduct)

					superGroup.POST("/banks/rename", questionHandler.RenameSource)
					superGroup.POST("/banks/delete", questionHandler.DeleteSource)
					superGroup.POST("/banks/transfer", questionHandler.TransferCategory)
					superGroup.POST("/categories/sync", questionHandler.SyncCategories)
					superGroup.PUT("/categories/:id", questionHandler.UpdateCategory)
					superGroup.POST("/categories/reorder", questionHandler.ReorderCategories)
					superGroup.POST("/questions/import", questionHandler.ImportQuestions)
					superGroup.PUT("/questions/:id", questionHandler.UpdateQuestion)
					superGroup.DELETE("/questions/:id", questionHandler.DeleteQuestion)
					superGroup.POST("/questions/batch-delete", questionHandler.BatchDeleteQuestions)
					superGroup.DELETE("/questions/by-category", questionHandler.DeleteByCategory)

					superGroup.POST("/forum/boards", forumHandler.CreateBoard)
					superGroup.PUT("/forum/boards/:id", forumHandler.UpdateBoard)
					superGroup.DELETE("/forum/boards/:id", forumHandler.DeleteBoard)

					superGroup.POST("/withdraw/handle", userHandler.HandleWithdraw)
					
					// 🔥🔥🔥 [新增] 删除接口 🔥🔥🔥
					superGroup.DELETE("/withdraw/:id", userHandler.DeleteWithdraw)
					superGroup.DELETE("/withdraw/clear", userHandler.ClearHandledWithdraws)
				}
			}
		}
	}
	return r
}