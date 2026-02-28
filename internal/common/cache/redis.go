package cache

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"med-platform/internal/common/db"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var RDB *redis.Client
var ctx = context.Background()

const PostViewKey = "forum:post:views"

// InitRedis 初始化 Redis 连接并启动后台同步任务
func InitRedis() {
	RDB = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // 默认本地 Redis 地址
		Password: "",               // 密码（如有请填写）
		DB:       0,                // 默认数据库
	})

	_, err := RDB.Ping(ctx).Result()
	if err != nil {
		log.Printf("⚠️ Redis 连接失败，浏览量将回退到直接写库模式: %v\n", err)
		RDB = nil
		return
	}
	fmt.Println("✅ Redis 初始化成功，已开启浏览量缓存机制")

	// 启动后台定时同步任务（不阻塞主线程）
	go syncViewsToMySQL()
}

// IncrPostView 增加帖子浏览量
func IncrPostView(postID uint) {
	if RDB != nil {
		// 写入 Redis Hash 表，键为帖子 ID，值为递增的浏览量
		RDB.HIncrBy(ctx, PostViewKey, strconv.Itoa(int(postID)), 1)
	} else {
		// 如果 Redis 没启动，作为降级方案直接写库
		db.DB.Table("forum_posts").Where("id = ?", postID).
			UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))
	}
}

// syncViewsToMySQL 定时将 Redis 中的浏览量同步到 MySQL
func syncViewsToMySQL() {
	// 设置定时器，每 5 分钟执行一次
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if RDB == nil {
			return
		}

		// 获取过去 5 分钟内所有被浏览过的帖子增量
		views, err := RDB.HGetAll(ctx, PostViewKey).Result()
		if err != nil || len(views) == 0 {
			continue
		}

		// 使用事务批量更新到 MySQL
		err = db.DB.Transaction(func(tx *gorm.DB) error {
			for postIDStr, countStr := range views {
				postID, _ := strconv.Atoi(postIDStr)
				count, _ := strconv.Atoi(countStr)

				if count > 0 {
					// 注意：这里使用 tx.Table("forum_posts") 而不是 tx.Model(&forum.ForumPost{})
					// 是为了避免引用 internal/forum 包从而导致循环引用报错
					tx.Table("forum_posts").Where("id = ?", postID).
						UpdateColumn("view_count", gorm.Expr("view_count + ?", count))
				}
			}
			return nil
		})

		// 同步成功后，清空当前的 Redis Hash 计数器
		if err == nil {
			RDB.Del(ctx, PostViewKey)
			fmt.Println("🔄 [Cron] 帖子浏览量缓冲池已成功批量写入 MySQL")
		} else {
			fmt.Printf("❌ [Cron] 浏览量同步 MySQL 失败: %v\n", err)
		}
	}
}