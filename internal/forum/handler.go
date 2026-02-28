package forum

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"med-platform/internal/common/cache"   // 🔥 引入新建的 cache 包
	"med-platform/internal/common/db"
	"med-platform/internal/common/model"   
	"med-platform/internal/common/service" 
	"med-platform/internal/common/uploader"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// =======================
// 1. 板块管理 (Board)
// =======================

func (h *Handler) CreateBoard(c *gin.Context) {
	var req ForumBoard
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Icon != "" && strings.Contains(req.Icon, "/uploads/temp/") {
		finalPaths := uploader.ConfirmImages([]string{req.Icon}, "forum")
		if len(finalPaths) > 0 {
			req.Icon = finalPaths[0]
		}
	}

	if err := db.DB.Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建板块失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "板块创建成功", "data": req})
}

func (h *Handler) ListBoards(c *gin.Context) {
	type BoardWithStats struct {
		ForumBoard
		PostCount    int64 `json:"post_count"`
		CommentCount int64 `json:"comment_count"`
	}

	var response []BoardWithStats
	err := db.DB.Table("forum_boards").
		Select("forum_boards.*, COUNT(DISTINCT forum_posts.id) as post_count, COALESCE(SUM(forum_posts.comment_count), 0) as comment_count").
		Joins("LEFT JOIN forum_posts ON forum_posts.board_id = forum_boards.id AND forum_posts.deleted_at IS NULL").
		Where("forum_boards.deleted_at IS NULL").
		Group("forum_boards.id").
		Order("sort_order desc, id asc").
		Scan(&response).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取板块失败"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"data": response})
}

func (h *Handler) UpdateBoard(c *gin.Context) {
	id := c.Param("id")
	var req ForumBoard
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var board ForumBoard
	if err := db.DB.First(&board, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "板块不存在"})
		return
	}

	if req.Icon != "" && strings.Contains(req.Icon, "/uploads/temp/") {
		finalPaths := uploader.ConfirmImages([]string{req.Icon}, "forum")
		if len(finalPaths) > 0 {
			req.Icon = finalPaths[0]
		}
	}

	req.ID = board.ID
	db.DB.Model(&board).Select("*").Updates(req)
	c.JSON(http.StatusOK, gin.H{"message": "更新成功", "data": board})
}

func (h *Handler) DeleteBoard(c *gin.Context) {
	id := c.Param("id")
	var count int64
	db.DB.Model(&ForumPost{}).Where("board_id = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该板块下还有帖子，无法删除"})
		return
	}
	db.DB.Delete(&ForumBoard{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "板块已删除"})
}

// =======================
// 2. 帖子管理 (Post)
// =======================

func extractSummary(htmlContent string, length int) string {
	re := regexp.MustCompile(`<[^>]*>`)
	plainText := re.ReplaceAllString(htmlContent, "")
	plainText = strings.ReplaceAll(plainText, "&nbsp;", " ")
	plainText = strings.ReplaceAll(plainText, "\n", "")
	plainText = strings.TrimSpace(plainText)
	
	runes := []rune(plainText)
	if len(runes) > length {
		return string(runes[:length]) + "..."
	}
	return string(runes)
}

func (h *Handler) CreatePost(c *gin.Context) {
	var req struct {
		BoardID  uint   `json:"board_id" binding:"required"`
		Title    string `json:"title" binding:"required"`
		Content  string `json:"content" binding:"required"`
		Summary  string `json:"summary"`
		IsPinned bool   `json:"is_pinned"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("userID").(uint)
	role := c.MustGet("role").(string)

	var board ForumBoard
	if err := db.DB.First(&board, req.BoardID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "板块不存在"})
		return
	}

	if board.IsLocked && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "该板块仅限管理员发帖"})
		return
	}

	re := regexp.MustCompile(`src="([^"]*\/uploads\/temp\/[^"]+)"`)
	matches := re.FindAllStringSubmatch(req.Content, -1)
	var relativePaths []string
	var originalUrls []string
	for _, match := range matches {
		fullUrl := match[1]
		if idx := strings.Index(fullUrl, "/uploads/temp/"); idx != -1 {
			relativePaths = append(relativePaths, fullUrl[idx:])
			originalUrls = append(originalUrls, fullUrl)
		}
	}
	if len(relativePaths) > 0 {
		finalPaths := uploader.ConfirmImages(relativePaths, "forum")
		for i, oldUrl := range originalUrls {
			req.Content = strings.Replace(req.Content, oldUrl, finalPaths[i], -1)
		}
	}

	summary := req.Summary
	if summary == "" {
		summary = extractSummary(req.Content, 100)
	}

	post := ForumPost{
		BoardID:  req.BoardID,
		AuthorID: userID,
		Title:    req.Title,
		Content:  req.Content,
		Summary:  summary,
		IsPinned: req.IsPinned && role == "admin",
	}

	if err := db.DB.Create(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发布失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "发布成功", "data": post})
}

func (h *Handler) ListPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	boardID := c.Query("board_id")
	keyword := c.Query("q")

	query := db.DB.Model(&ForumPost{}).Preload("Author").Preload("Board").Order("is_global desc, is_pinned desc, created_at desc")
	if boardID != "" {
		query = query.Where("board_id = ?", boardID)
	}
	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)
	var posts []ForumPost
	query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&posts)

	c.JSON(http.StatusOK, gin.H{"data": posts, "total": total, "page": page})
}

func (h *Handler) GetPostDetail(c *gin.Context) {
	id := c.Param("id")
	var post ForumPost
	if err := db.DB.Preload("Author").Preload("Board").First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
		return
	}
	
	// 🔥🔥🔥 核心修改：调用 Redis 进行高频缓冲，而不是直接 UpdateColumn
	cache.IncrPostView(post.ID)
	post.ViewCount++ // 让前端立刻看到 +1 效果
	
	c.JSON(http.StatusOK, gin.H{"data": post})
}

func (h *Handler) DeletePost(c *gin.Context) {
    postID, _ := strconv.Atoi(c.Param("id"))
    userID := c.MustGet("userID").(uint)
    userRole := c.MustGet("role").(string)

    var post ForumPost // 直接引用当前包下的结构体
    if err := db.DB.First(&post, postID).Error; err != nil {
        c.JSON(404, gin.H{"error": "帖子不存在"})
        return
    }

    // 2. 权限校验：既不是管理员，也不是原作者
    if userRole != "admin" && post.AuthorID != userID {
        c.JSON(403, gin.H{"error": "你没有权限删除他人的帖子"})
        return
    }

    // 3. 执行删除
    if err := db.DB.Delete(&post).Error; err != nil {
        c.JSON(500, gin.H{"error": "删除失败"})
        return
    }
    c.JSON(200, gin.H{"message": "已成功删除"})
}

func (h *Handler) UploadImage(c *gin.Context) {
	url, err := uploader.SaveImageWithHash(c, "file", 5*1024*1024)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "上传成功", "url": url, "data": map[string]string{"url": url}})
}

// =======================
// 4. 评论与互动 (Comment & Report)
// =======================

func (h *Handler) CreateComment(c *gin.Context) {
	var req struct {
		PostID   uint   `json:"post_id" binding:"required"`
		Content  string `json:"content" binding:"required"`
		ParentID *uint  `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	userID := c.MustGet("userID").(uint)

	var post ForumPost
	if err := db.DB.First(&post, req.PostID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
		return
	}

	actualParentID := req.ParentID
	if req.ParentID != nil {
		var parentComment ForumComment
		if err := db.DB.Select("id, parent_id, author_id").First(&parentComment, *req.ParentID).Error; err == nil {
			if parentComment.ParentID != nil {
				actualParentID = parentComment.ParentID
			}
		}
	}

	re := regexp.MustCompile(`src="([^"]*\/uploads\/temp\/[^"]+)"`)
	matches := re.FindAllStringSubmatch(req.Content, -1)
	var relativePaths []string
	var originalUrls []string
	for _, match := range matches {
		fullUrl := match[1]
		if idx := strings.Index(fullUrl, "/uploads/temp/"); idx != -1 {
			relativePaths = append(relativePaths, fullUrl[idx:])
			originalUrls = append(originalUrls, fullUrl)
		}
	}
	if len(relativePaths) > 0 {
		finalPaths := uploader.ConfirmImages(relativePaths, "forum")
		for i, oldUrl := range originalUrls {
			req.Content = strings.Replace(req.Content, oldUrl, finalPaths[i], -1)
		}
	}

	comment := ForumComment{
		PostID:   req.PostID,
		AuthorID: userID,
		Content:  req.Content,
		ParentID: actualParentID,
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&comment).Error; err != nil {
			return err
		}
		return tx.Model(&post).UpdateColumn("comment_count", gorm.Expr("comment_count + ?", 1)).Error
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "评论提交失败"})
		return
	}

	var targetUserID uint = post.AuthorID 
	if req.ParentID != nil && *req.ParentID > 0 {
		var parentComment ForumComment
		if err := db.DB.Select("author_id").First(&parentComment, *req.ParentID).Error; err == nil {
			targetUserID = parentComment.AuthorID
		}
	}

	service.SendNotification(
		targetUserID,
		userID,
		"forum",
		post.ID,
		req.Content, 
		post.Title,  
	)

	c.JSON(http.StatusOK, gin.H{"message": "评论成功"})
}

func (h *Handler) ListComments(c *gin.Context) {
	postID := c.Query("post_id")
	var comments []ForumComment
	
	db.DB.Where("post_id = ? AND parent_id IS NULL", postID).
		Preload("Author").
		Preload("Children.Author").
		Order("created_at asc").
		Find(&comments)
		
	c.JSON(http.StatusOK, gin.H{"data": comments})
}

func (h *Handler) DeleteComment(c *gin.Context) {
    commentID, _ := strconv.Atoi(c.Param("id"))
    userID := c.MustGet("userID").(uint)
    userRole := c.MustGet("role").(string)

    var comment ForumComment // 直接引用当前包下的结构体
    if err := db.DB.First(&comment, commentID).Error; err != nil {
        c.JSON(404, gin.H{"error": "评论不存在"})
        return
    }

    // 权限校验
    if userRole != "admin" && comment.AuthorID != userID {
        c.JSON(403, gin.H{"error": "你没有权限删除他人的评论"})
        return
    }

    if err := db.DB.Delete(&comment).Error; err != nil {
        c.JSON(500, gin.H{"error": "删除失败"})
        return
    }
    c.JSON(200, gin.H{"message": "已成功删除"})
}

func (h *Handler) CreateReport(c *gin.Context) {
	var req struct {
		TargetID   uint   `json:"target_id" binding:"required"`
		TargetType string `json:"target_type" binding:"required"`
		Reason     string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("userID").(uint)
	report := model.ForumReport{
		TargetID:   req.TargetID,
		TargetType: req.TargetType,
		Reason:     req.Reason,
		ReporterID: userID,
		Status:     0,
	}
	db.DB.Create(&report)
	c.JSON(http.StatusOK, gin.H{"message": "举报已提交"})
}

// =======================
// 5. 管理员专属接口
// =======================

func (h *Handler) AdminListComments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("q")

	query := db.DB.Model(&ForumComment{}).Preload("Author").Order("created_at desc")
	if keyword != "" {
		query = query.Where("content LIKE ?", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)
	var comments []ForumComment
	query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&comments)

	c.JSON(http.StatusOK, gin.H{"data": comments, "total": total, "page": page})
}

func (h *Handler) AdminListReports(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")

	query := db.DB.Model(&model.ForumReport{}).Preload("Reporter").Order("status asc, created_at desc")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)
	var reports []model.ForumReport
	query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&reports)

	c.JSON(http.StatusOK, gin.H{"data": reports, "total": total})
}

func (h *Handler) AdminResolveReport(c *gin.Context) {
	id := c.Param("id")
	if err := db.DB.Model(&model.ForumReport{}).Where("id = ?", id).Update("status", 1).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "举报已标记为处理"})
}

func (h *Handler) AdminGetReportContent(c *gin.Context) {
	targetType := c.Query("target_type")
	targetID := c.Query("target_id")

	if targetType == "post" {
		var post ForumPost
		if err := db.DB.Preload("Author").First(&post, targetID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "帖子已不存在"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"content": post.Content,
			"author":  post.Author.Nickname,
			"title":   post.Title,
			"type":    "post",
		})
	} else if targetType == "comment" {
		var comment ForumComment
		if err := db.DB.Preload("Author").First(&comment, targetID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "评论已不存在"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"content": comment.Content,
			"author":  comment.Author.Nickname,
			"title":   "评论内容",
			"type":    "comment",
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未知类型"})
	}
}

// =======================
// 6. 通用消息通知接口 
// =======================

func (h *Handler) GetNotifications(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	var notifs []model.Notification

	err := db.DB.Where("user_id = ?", userID).
		Preload("Sender"). 
		Order("is_read asc, created_at desc").
		Limit(20).
		Find(&notifs).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取消息失败"})
		return
	}

	var unreadCount int64
	db.DB.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&unreadCount)

	c.JSON(http.StatusOK, gin.H{
		"data":         notifs, 
		"unread_count": unreadCount,
	})
}

func (h *Handler) ReadNotification(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	id := c.Param("id")

	db.DB.Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true)

	c.JSON(http.StatusOK, gin.H{"message": "已读"})
}

func (h *Handler) ReadAllNotifications(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	db.DB.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true)

	c.JSON(http.StatusOK, gin.H{"message": "全部已读"})
}