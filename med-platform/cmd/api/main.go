package main

import (
	"fmt"
	"med-platform/internal/sysconfig"
	"med-platform/internal/answer"
	"med-platform/internal/common/config"
	"med-platform/internal/common/cron"
	"med-platform/internal/common/db"
	"med-platform/internal/common/logger"
	"med-platform/internal/common/model"
	"med-platform/internal/common/service" // 🔥 修复：补全 service 包导入
	"med-platform/internal/feedback"
	"med-platform/internal/forum"
	"med-platform/internal/note"
	"med-platform/internal/payment" 
	"med-platform/internal/common/cache"
	"med-platform/internal/product"
	"med-platform/internal/question"
	"med-platform/internal/router"
	"med-platform/internal/user"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 1. 初始化
	config.Load()
	logger.Init(config.GlobalConfig.App.Env)
	defer logger.Log.Sync()
	db.Init()
	logger.Log.Info("database connected successfully")
    
	// 1.1 初始化缓存与实时推送枢纽
	cache.InitRedis()
	go service.Hub.Run() // 🔥 现在这里不会报错 undefined 了

	// 2. 数据库迁移
	err := db.DB.AutoMigrate(
		&user.User{},
		&user.VerificationToken{},
		&question.Question{},
		&question.Category{},
		&question.UserDailyStat{},
		&question.UserArchivedStat{},
		&question.QuestionFeedback{},
		
		&answer.AnswerRecord{},
		&answer.UserMistake{},
		&answer.UserFavorite{},
		&answer.AnswerHistory{}, 
		
		&note.Note{},
		&note.NoteLike{},
		&note.NoteCollect{},
		&note.NoteReport{},
		
		&product.Product{},
		&product.ProductSku{},
		&product.ProductContent{},
		&product.UserProduct{},
		&product.ProductAuthLog{},
		&product.ExchangeRecord{},

		&payment.Order{},           
		&payment.CommissionLog{},   
		&payment.WithdrawRequest{}, 
		&payment.ActivationCode{},  

		&feedback.PlatformFeedback{},
		&forum.ForumBoard{},
		&forum.ForumPost{},
		&forum.ForumComment{},
		&model.ForumReport{},  // 🔥 修复：改为 model.ForumReport，解决 undefined 报错
		&model.Notification{}, // 统一使用 model 包下的通知模型
        
		&sysconfig.SysConfig{},
	)
	if err != nil {
		logger.Log.Fatal("database migration failed", zap.Error(err))
	}

	sysconfig.InitConfig()

	// 3. 启动任务
	fmt.Println("正在校准目录树数据...")
	question.NewRepository().SyncCategories()
	
	fmt.Println("正在启动数据归档任务...")
	go answer.StartArchivingTask()
	cron.StartBackgroundTasks()

	// 4. 启动服务
	if config.GlobalConfig.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := router.SetupRouter()
	addr := fmt.Sprintf(":%d", config.GlobalConfig.App.Port)
	logger.Log.Info("Server running on " + addr)
	r.Run(addr)
}