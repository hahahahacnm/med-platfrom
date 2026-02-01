package service

import (
	"med-platform/internal/common/db"
	"med-platform/internal/common/model"
)

// SendNotification 发送通用通知
// targetUserID: 接收者
// senderID: 发送者
// sourceType: "forum" 或 "question"
// sourceID: 帖子ID 或 题目ID
// content: 评论内容
// title: 来源标题（帖子名或题目摘要）
func SendNotification(targetUserID, senderID uint, sourceType string, sourceID uint, content, title string) {
	// 1. 自己不通知自己
	if targetUserID == senderID {
		return
	}

	// 2. 截取摘要，防止太长
	summary := content
	if len([]rune(summary)) > 30 {
		summary = string([]rune(summary)[:30]) + "..."
	}
    
    // 3. 构造通知
	notif := model.Notification{
		UserID:     targetUserID,
		SenderID:   senderID,
		SourceType: sourceType,
		SourceID:   sourceID,
		Content:    summary,
		Title:      title,
		IsRead:     false,
	}

	// 4. 异步入库 (不阻塞主线程)
	go db.DB.Create(&notif)
}