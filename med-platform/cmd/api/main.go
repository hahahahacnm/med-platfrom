package main

import (
	"fmt"
	"med-platform/internal/answer"
	"med-platform/internal/common/config"
	"med-platform/internal/common/cron"
	"med-platform/internal/common/db"
	"med-platform/internal/common/logger"
	"med-platform/internal/common/model" // 🔥 1. 引入通用模型包 (存放 Notification)
	"med-platform/internal/feedback"
	"med-platform/internal/forum"
	"med-platform/internal/note"
	"med-platform/internal/payment"
	"med-platform/internal/product"
	"med-platform/internal/question"
	"med-platform/internal/router"
	"med-platform/internal/user"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 1️⃣ 基础建设初始化
	config.Load()
	logger.Init(config.GlobalConfig.App.Env)
	defer logger.Log.Sync()
	db.Init()
	logger.Log.Info("database connected successfully")

	// 2️⃣ 数据库迁移 (统一管理所有表结构)
	err := db.DB.AutoMigrate(
		&user.User{},
		&question.Question{},
		&question.Category{},
		&question.UserDailyStat{},    // 每日统计 (热数据)
		&question.UserArchivedStat{}, // 历史归档 (冷数据)
		&question.QuestionFeedback{},
		&answer.AnswerRecord{},
		&answer.UserMistake{},
		&answer.UserFavorite{},
		&note.Note{},
		&note.NoteLike{},
		&note.NoteCollect{},
		&note.NoteReport{},
		&product.Product{},
		&product.ProductContent{},
		&product.ProductSku{},
		&product.UserProduct{},
		&product.ProductAuthLog{},
		&payment.Order{},
		&feedback.PlatformFeedback{},
		&product.SalesRecord{},       // 销售记录
		&product.WithdrawRequest{},   // 🔥🔥🔥 [新增] 提现申请表，解决 42P01 错误 🔥🔥🔥

		// 论坛/社区模块
		&forum.ForumBoard{},
		&forum.ForumPost{},
		&forum.ForumComment{},
		&forum.ForumReport{},

		// 🔥🔥🔥 2. 新增：全局消息通知表 (打通论坛与题库) 🔥🔥🔥
		&model.Notification{},
	)
	if err != nil {
		logger.Log.Fatal("database migration failed", zap.Error(err))
	}

	// 校准目录树
	fmt.Println("正在校准目录树数据...")
	question.NewRepository().SyncCategories()
	fmt.Println("✅ 目录树校准完成")

	// 3️⃣ 启动后台任务
	// 启动数据归档守护进程
	fmt.Println("正在启动数据归档任务...")
	go answer.StartArchivingTask()
	fmt.Println("✅ 数据归档任务已后台运行")

	// 启动临时文件清理任务
	cron.StartBackgroundTasks()


	// 4️⃣ 启动服务
	if config.GlobalConfig.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 使用 router 包来统一管理路由
	r := router.SetupRouter()

	addr := fmt.Sprintf(":%d", config.GlobalConfig.App.Port)
	logger.Log.Info("Server running on " + addr)
	r.Run(addr)
}