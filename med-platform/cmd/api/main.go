package main

import (
	"fmt"
	"med-platform/internal/answer"
	"med-platform/internal/common/config"
	"med-platform/internal/common/db"
	"med-platform/internal/common/logger"
	"med-platform/internal/note"
	"med-platform/internal/product"
	"med-platform/internal/question"
	"med-platform/internal/router" // 👈 引入 router 包
	"med-platform/internal/user"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 1️⃣ 基础建设初始化
	config.Load()
	logger.Init(config.Cfg.App.Env)
	defer logger.Log.Sync()
	db.Init()
	logger.Log.Info("database connected successfully")

	// 2️⃣ 数据库迁移 (🔥 新增 UserDailyStat 和 UserArchivedStat)
	err := db.DB.AutoMigrate(
		&user.User{},
		&question.Question{},
		&question.Category{},
		&question.UserDailyStat{},    // 🔥 新增：每日统计表 (热数据)
		&question.UserArchivedStat{}, // 🔥 新增：历史归档表 (冷数据)
		&answer.AnswerRecord{},
		&answer.UserMistake{},
		&answer.UserFavorite{},
		&note.Note{},
		&note.NoteLike{},
		&note.NoteCollect{},
		&product.Product{},
		&product.ProductContent{},
		&product.UserProduct{},
		&product.ProductAuthLog{},
	)
	if err != nil {
		logger.Log.Fatal("database migration failed", zap.Error(err))
	}

	// 校准目录树
	fmt.Println("正在校准目录树数据...")
	question.NewRepository().SyncCategories()
	fmt.Println("✅ 目录树校准完成")

	// 3️⃣ 启动后台任务 (🔥 新增)
	// 启动数据归档守护进程，每天自动把超过1年的数据搬到冷库
	fmt.Println("正在启动数据归档任务...")
	go answer.StartArchivingTask()
	fmt.Println("✅ 数据归档任务已后台运行")

	// 4️⃣ 启动服务
	if config.Cfg.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 使用 router 包来统一管理路由
	r := router.SetupRouter()

	addr := fmt.Sprintf(":%d", config.Cfg.App.Port)
	logger.Log.Info("Server running on " + addr)
	r.Run(addr)
}