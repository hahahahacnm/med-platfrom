package db

import (
	"log"
	"med-platform/internal/common/config" // 注意替换你的模块名
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	var err error
	dsn := config.GlobalConfig.Data.Database.Source

	// 打开连接
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect db failed: %v", err)
	}

	// 设置连接池（这是生产环境必须的配置）
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("get db instance failed: %v", err)
	}

	// 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(10)
	// 设置打开数据库连接的最大数量
	sqlDB.SetMaxOpenConns(100)
	// 设置了连接可复用的最大时间
	sqlDB.SetConnMaxLifetime(time.Hour)
}