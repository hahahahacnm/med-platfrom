package service

import (
	"fmt"
	"html"
	"med-platform/internal/common/db"
	"med-platform/internal/common/model"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// SendNotification 发送通用通知并实现 WebSocket 实时推送
func SendNotification(targetUserID, senderID uint, sourceType string, sourceID uint, content, title string) {
	// 1. 基础校验：自己不通知自己，或者目标 ID 非法时退出
	if targetUserID == senderID || targetUserID == 0 {
		return
	}

	// 2. 内容预处理：清洗 HTML 标签并还原转义字符（如 &nbsp; 还原为空格）
	summary := cleanHTML(content)
	
	// 3. 长度截断：保持通知精简
	runes := []rune(summary)
	if len(runes) > 40 {
		summary = string(runes[:40]) + "..."
	}
    
	// 4. 构造通知模型
	notif := model.Notification{
		UserID:     targetUserID,
		SenderID:   senderID,
		SourceType: sourceType,
		SourceID:   sourceID,
		Content:    summary,
		Title:      title,
		IsRead:     false,
	}

	// 5. 异步执行入库与实时推送
	go func() {
		// 存储到 MySQL 数据库
		if err := db.DB.Create(&notif).Error; err != nil {
			fmt.Printf("[Notify Error] 数据库写入失败: %v\n", err)
			return
		}

		// 6. 实时触达：通过 WebSocket 发送消息给在线用户
		// 使用我们第二阶段创建的全局 Hub
		if Hub != nil {
			Hub.SendToUser(targetUserID, gin.H{
				"type": "new_notification",
				"data": gin.H{
					"id":          notif.ID,
					"title":       notif.Title,
					"content":     notif.Content,
					"source_type": notif.SourceType,
					"source_id":   notif.SourceID,
					"created_at":  notif.CreatedAt,
				},
			})
			fmt.Printf("📢 [WebSocket] 已向用户 %d 推送实时通知\n", targetUserID)
		}
	}()
}

// cleanHTML 辅助函数：彻底清洗 HTML 标签、多余空格和转义字符
func cleanHTML(input string) string {
	// 移除所有 <...> 标签
	re := regexp.MustCompile(`<[^>]*>`)
	output := re.ReplaceAllString(input, "")
	
	// 处理 HTML 转义字符 (如 &nbsp; 变回空格)
	output = html.UnescapeString(output)
	
	// 将多个连续空格或换行符替换为单个空格
	output = strings.ReplaceAll(output, "\n", " ")
	output = strings.TrimSpace(output)
	
	return output
}