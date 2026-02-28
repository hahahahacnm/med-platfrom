package router

import (
	"med-platform/internal/answer"
	"med-platform/internal/common/captcha"
	"med-platform/internal/common/middleware"
	"med-platform/internal/common/service"
	"med-platform/internal/feedback"
	"med-platform/internal/forum"
	"med-platform/internal/note"
	"med-platform/internal/payment"
	"med-platform/internal/product"
	"med-platform/internal/question"
	"med-platform/internal/sysconfig"
	"med-platform/internal/user"

	"github.com/gin-gonic/gin"
)

// RouteManager 路由管理器：持有所有 Handler 和限流器
type RouteManager struct {
	// Handlers
	user      *user.Handler
	question  *question.Handler
	answer    *answer.Handler
	note      *note.Handler
	product   *product.Handler
	payment   *payment.Handler
	feedback  *feedback.Handler
	forum     *forum.Handler
	sysconfig *sysconfig.Handler

	// Limiters (限流器)
	commentLimiter *middleware.IPRateLimiter
	uploadLimiter  *middleware.IPRateLimiter
}

// SetupRouter 初始化路由入口
func SetupRouter() *gin.Engine {
	// 1. 初始化基础服务
	captcha.Init()

	// 2. 初始化 Gin 引擎
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	r.Static("/uploads", "./uploads")
	r.GET("/ws", service.WsHandler)

	// 3. 构建路由管理器 (初始化所有 Handler)
	mgr := &RouteManager{
		user:      user.NewHandler(),
		question:  question.NewHandler(),
		answer:    answer.NewHandler(),
		note:      &note.Handler{},
		product:   product.NewHandler(),
		payment:   payment.NewHandler(),
		feedback:  feedback.NewHandler(),
		forum:     forum.NewHandler(),
		sysconfig: sysconfig.NewHandler(),

		// 针对不同场景的限流策略
		commentLimiter: middleware.NewIPRateLimiter(1, 3), // 发言：1秒3次
		uploadLimiter:  middleware.NewIPRateLimiter(2, 5), // 上传：2秒5次
	}

	// 4. 注册路由组
	rootGroup := r.Group("/api/v1")
	{
		mgr.registerPublicRoutes(rootGroup) // 🟢 公共接口 (无需登录)
		mgr.registerAuthRoutes(rootGroup)   // 🟠 需登录接口 (JWT)
	}

	return r
}

// 🟢 注册公共接口
func (m *RouteManager) registerPublicRoutes(g *gin.RouterGroup) {
	// 验证码
	g.GET("/auth/captcha", m.user.GetCaptcha)

	// 认证与基础
	g.POST("/auth/register", m.user.Register)
	g.POST("/auth/login", m.user.Login)

	// 🔥 修复点：已将 /category-tree 移出公共路由，移动到下方的 registerQuestionRoutes 中

	// 魔法链接相关接口
	g.GET("/auth/verify-email", m.user.VerifyEmail)
	g.POST("/auth/resend-email", m.user.ResendEmail)

	// 支付回调
	g.GET("/payment/mock/callback", m.payment.MockSuccess)
}

// 🟠 注册认证接口 (需 JWT Token)
func (m *RouteManager) registerAuthRoutes(parent *gin.RouterGroup) {
	userGroup := parent.Group("/")
	userGroup.Use(middleware.AuthJWT())

	// === 业务模块挂载 ===
	m.registerUploadRoutes(userGroup)
	m.registerForumRoutes(userGroup)
	m.registerUserCenterRoutes(userGroup)
	m.registerQuestionRoutes(userGroup)
	m.registerNoteRoutes(userGroup)
	m.registerCommerceRoutes(userGroup)

	// === 后台管理模块 (内部鉴权) ===
	m.registerAdminRoutes(userGroup)
}

// 📁 上传模块
func (m *RouteManager) registerUploadRoutes(g *gin.RouterGroup) {
	limit := middleware.RateLimitMiddleware(m.uploadLimiter)
	g.POST("/upload/avatar", limit, m.user.UploadAvatar)
	g.POST("/upload/payment", limit, m.user.UploadPaymentCode)
	g.POST("/upload", limit, m.user.UploadAvatar)
}

// 💬 论坛与消息模块
func (m *RouteManager) registerForumRoutes(g *gin.RouterGroup) {
	// 消息通知
	g.GET("/notifications", m.forum.GetNotifications)
	g.PUT("/notifications/:id/read", m.forum.ReadNotification)
	g.PUT("/notifications/read-all", m.forum.ReadAllNotifications)

	// 帖子与评论
	g.GET("/forum/boards", m.forum.ListBoards)
	g.GET("/forum/posts", m.forum.ListPosts)
	g.GET("/forum/posts/:id", m.forum.GetPostDetail)
	g.POST("/forum/posts", m.forum.CreatePost)
	g.DELETE("/forum/posts/:id", m.forum.DeletePost)

	// 图片上传 (带限流)
	g.POST("/forum/upload", middleware.RateLimitMiddleware(m.uploadLimiter), m.forum.UploadImage)

	g.GET("/forum/comments", m.forum.ListComments)
	g.POST("/forum/comments", m.forum.CreateComment)
	g.DELETE("/forum/comments/:id", m.forum.DeleteComment)
	g.POST("/forum/report", m.forum.CreateReport)
}

// 👤 个人中心模块
func (m *RouteManager) registerUserCenterRoutes(g *gin.RouterGroup) {
	g.GET("/user/profile", m.user.GetProfile)
	g.PUT("/user/profile", m.user.UpdateProfile)
	g.POST("/user/avatar", m.user.UploadAvatar)
	g.PUT("/user/password", m.user.ChangePassword)

	g.POST("/user/email/bind", m.user.BindNewEmail)
}

// 📚 题库与答题模块
func (m *RouteManager) registerQuestionRoutes(g *gin.RouterGroup) {
	// 题目基础
	// 🔥 修复点：将目录树接口移入认证路由，以便解析 userID 统计进度
	g.GET("/category-tree", m.question.GetTree)
	g.GET("/questions/skeleton", m.question.GetChapterSkeleton)
	g.GET("/questions", m.question.List)
	g.GET("/questions/:id", m.question.GetDetail)
	g.GET("/banks", m.question.GetSources)
	g.POST("/questions/:id/submit", m.answer.Submit)
	g.POST("/feedback", m.question.SubmitFeedback)

	// 错题/收藏/统计
	g.GET("/mistakes/skeleton", m.answer.GetMistakeSkeleton)
	g.GET("/favorites/skeleton", m.answer.GetFavoriteSkeleton)
	g.GET("/mistakes", m.answer.GetMistakes)
	g.DELETE("/mistakes/:id", m.answer.RemoveMistake)
	g.GET("/mistake-tree", m.answer.GetMistakeTree)
	g.GET("/stats", m.answer.GetStats)
	g.GET("/rank/daily", m.answer.GetDailyRank)
	g.POST("/favorites/:id", m.answer.ToggleFavorite)
	g.GET("/favorites", m.answer.GetFavorites)
	g.GET("/favorite-tree", m.answer.GetFavoriteTree)

	// 重置与清理
	g.DELETE("/questions/:id/reset", m.answer.Reset)
	g.DELETE("/answers/reset-chapter", m.answer.ResetChapter)
}

// 📝 笔记模块
func (m *RouteManager) registerNoteRoutes(g *gin.RouterGroup) {
	limit := middleware.RateLimitMiddleware(m.commentLimiter)
	g.POST("/notes", limit, m.note.SaveNote)
	g.GET("/notes", m.note.ListNotes)
	g.GET("/notes/my", m.note.GetMyNotes)
	g.GET("/notes/skeleton", m.note.GetNoteSkeleton)
	g.GET("/notes/note-tree", m.note.GetNoteTree)
	g.DELETE("/notes/:id", m.note.DeleteNote)
	g.POST("/notes/upload", middleware.RateLimitMiddleware(m.uploadLimiter), m.note.UploadImage)
	g.POST("/notes/:id/like", m.note.ToggleLike)
	g.POST("/notes/:id/collect", m.note.ToggleCollect)
	g.POST("/notes/:id/report", m.note.ReportNote)
}

// 💳 商业化模块 (支付/商品/反馈)
func (m *RouteManager) registerCommerceRoutes(g *gin.RouterGroup) {
	g.GET("/market/products", m.product.ListMarketProducts)

	// 🔥 新增这一行：获取单个商品的详细信息（富文本详情等）
	g.GET("/market/products/:id", m.product.GetProductDetail)

	g.POST("/pay/create", m.payment.CreatePay)
	g.POST("/codes/redeem", m.payment.RedeemCode)

	g.POST("/product/exchange", m.product.ExchangeProduct)

	g.GET("/user/products/:id", m.product.GetUserProducts)
	g.POST("/platform-feedback", m.feedback.Create)
	g.GET("/platform-feedback", m.feedback.GetMyList)
}

// 🔴 注册后台管理接口
func (m *RouteManager) registerAdminRoutes(parent *gin.RouterGroup) {
	// 1️⃣ 员工组 (Staff)
	staffGroup := parent.Group("/admin")
	staffGroup.Use(middleware.RequireStaff())
	{
		// 商品管理
		staffGroup.GET("/products", m.product.ListProducts)
		staffGroup.GET("/products/:id/contents", m.product.GetProductContents)
		staffGroup.POST("/products/bind", m.product.BindContent)
		staffGroup.POST("/products/unbind", m.product.UnbindContent)
		staffGroup.POST("/users/grant", m.product.GrantProductToUser)
		staffGroup.POST("/users/revoke", m.product.RevokeUserProduct)

		// 用户查询
		staffGroup.GET("/users", m.user.ListUsers)
		staffGroup.GET("/users/:id", m.user.AdminGetUserDetail)
		staffGroup.GET("/users/:id/products", m.product.GetUserProducts)
		staffGroup.GET("/auth-logs", m.product.GetAuthLogs)

		// 财务与提现
		staffGroup.GET("/dashboard/stats", m.user.GetDashboardStats)
		staffGroup.POST("/withdraw/apply", m.user.ApplyWithdraw)

		// 内容审核
		staffGroup.GET("/forum/comments", m.forum.AdminListComments)
		staffGroup.DELETE("/forum/comments/:id", m.forum.DeleteComment)
		staffGroup.DELETE("/forum/posts/:id", m.forum.DeletePost)
		staffGroup.GET("/forum/reports", m.forum.AdminListReports)
		staffGroup.GET("/forum/reports/preview", m.forum.AdminGetReportContent)
		staffGroup.PUT("/forum/reports/:id/resolve", m.forum.AdminResolveReport)
		staffGroup.GET("/notes", m.note.AdminListNotes)
		staffGroup.POST("/notes/:id/ignore", m.note.AdminDismissReport)
		staffGroup.GET("/feedbacks", m.question.AdminListFeedbacks)
		staffGroup.PUT("/feedbacks/:id", m.question.AdminResolveFeedback)
		staffGroup.GET("/platform-feedbacks", m.feedback.AdminList)
		staffGroup.PUT("/platform-feedbacks/:id", m.feedback.AdminReply)

		// 2️⃣ 超级管理员组 (SuperAdmin)
		superGroup := staffGroup.Group("/")
		superGroup.Use(middleware.RequireSuperAdmin())
		{
			superGroup.GET("/configs", m.sysconfig.ListConfigs)
			superGroup.POST("/configs", m.sysconfig.SaveConfig)
			superGroup.POST("/configs/test-email", m.sysconfig.SendTestEmail)
			// 🔥 新增：邮件营销/群发后台
			superGroup.GET("/emails/users", m.sysconfig.ListEmailUsers)
			superGroup.POST("/emails/send", m.sysconfig.SendCustomMail)

			// 用户敏感操作
			superGroup.POST("/users/role", m.user.UpdateRole)
			superGroup.POST("/users/ban", m.user.BanUser)
			superGroup.POST("/users/unban", m.user.UnbanUser)
			superGroup.PUT("/users/:id", m.user.AdminUpdateUserInfo)
			superGroup.PUT("/users/:id/password", m.user.AdminResetPassword)
			superGroup.POST("/users/:id/avatar", m.user.AdminUploadAvatar)

			// 商品定义
			superGroup.POST("/products", m.product.CreateProduct)
			superGroup.PUT("/products/:id", m.product.UpdateProduct)
			superGroup.DELETE("/products/:id", m.product.DeleteProduct)
			superGroup.POST("/codes/generate", m.payment.GenerateCodes)
			superGroup.GET("/codes", m.payment.ListCodes)
			superGroup.GET("/codes/export", m.payment.ExportCodes)
			superGroup.POST("/users/points", m.payment.ManualUpdatePoints)
			superGroup.POST("/products/upload", middleware.RateLimitMiddleware(m.uploadLimiter), m.product.UploadCover)

			// 题库管理
			superGroup.POST("/banks/rename", m.question.RenameSource)
			superGroup.POST("/banks/delete", m.question.DeleteSource)
			superGroup.POST("/banks/transfer", m.question.TransferCategory)
			superGroup.POST("/categories/sync", m.question.SyncCategories)
			superGroup.PUT("/categories/:id", m.question.UpdateCategory)
			superGroup.POST("/categories/reorder", m.question.ReorderCategories)
			superGroup.POST("/questions/import", m.question.ImportQuestions)
			superGroup.PUT("/questions/:id", m.question.UpdateQuestion)
			superGroup.DELETE("/questions/:id", m.question.DeleteQuestion)
			superGroup.POST("/questions/batch-delete", m.question.BatchDeleteQuestions)
			superGroup.DELETE("/questions/by-category", m.question.DeleteByCategory)

			// 论坛板块
			superGroup.POST("/forum/boards", m.forum.CreateBoard)
			superGroup.PUT("/forum/boards/:id", m.forum.UpdateBoard)
			superGroup.DELETE("/forum/boards/:id", m.forum.DeleteBoard)

			// 提现审核与清理
			superGroup.POST("/withdraw/handle", m.user.HandleWithdraw)
			superGroup.DELETE("/withdraw/:id", m.user.DeleteWithdraw)
			superGroup.DELETE("/withdraw/clear", m.user.ClearHandledWithdraws)
		}
	}
}
