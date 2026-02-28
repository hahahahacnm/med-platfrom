package note

import (
	"med-platform/internal/question"
	"med-platform/internal/user"
	"time"

	"gorm.io/gorm"
)

// Note 笔记主表
type Note struct {
	ID             uint              `gorm:"primarykey" json:"id"`
	UserID         uint              `gorm:"index;not null" json:"user_id"`
	QuestionID     uint              `gorm:"index;not null" json:"question_id"`
	
	// 内容可能为空（如果是纯图片模式）
	Content        string            `gorm:"type:text" json:"content"`

	// 🔥🔥🔥 新增：图片列表 🔥🔥🔥
	// GORM 会自动把 []string 转成 json 字符串存入数据库
	// 前端传参时传 ["/uploads/1.jpg", "/uploads/2.jpg"]
	Images         []string          `gorm:"serializer:json" json:"images"`

	IsPublic       bool              `gorm:"default:true" json:"is_public"`
	ParentID       *uint             `gorm:"index" json:"parent_id"`

	// 统计数据
	LikeCount      int               `gorm:"default:0" json:"like_count"`

	// 🔥🔥🔥 新增：举报相关字段 🔥🔥🔥
	IsReported     bool              `gorm:"default:false;index" json:"is_reported"` // 是否进入举报列表
	ReportCount    int               `gorm:"default:0" json:"report_count"`          // 被举报次数

	// ================= 关联关系 =================
	User           user.User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Question       question.Question `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
	Parent         *Note             `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	
	// 🔥 方便管理员查看该笔记被举报的具体记录
	Reports        []NoteReport      `gorm:"foreignKey:NoteID" json:"reports"`

	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	DeletedAt      gorm.DeletedAt    `gorm:"index" json:"deleted_at,omitempty"` // 支持软删除

	// ================= 动态状态 (非数据库字段) =================
	IsLiked        bool              `gorm:"-" json:"is_liked"`
	IsCollected    bool              `gorm:"-" json:"is_collected"`
}

func (Note) TableName() string {
	return "notes"
}

// NoteLike 点赞记录表
type NoteLike struct {
	ID        uint      `gorm:"primaryKey"`
	// 这里去掉了 uniqueIndex，配合 Handler 里的 GetStartOfDay 逻辑，
	// 实现了“每天可以点赞一次”或“取消后再点”的宽松逻辑。
	UserID    uint      `gorm:"index" json:"user_id"`
	NoteID    uint      `gorm:"index" json:"note_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (NoteLike) TableName() string {
	return "note_likes"
}

// NoteCollect 收藏表 (收藏是持久状态，所以必须唯一)
type NoteCollect struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"uniqueIndex:idx_user_note_collect" json:"user_id"`
	NoteID    uint      `gorm:"uniqueIndex:idx_user_note_collect" json:"note_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (NoteCollect) TableName() string { return "note_collects" }

// NoteReport 举报记录表 (防止重复举报 + 记录理由)
type NoteReport struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"uniqueIndex:idx_user_note_report" json:"user_id"` // 联合唯一索引：一人一贴只能举报一次
	NoteID    uint      `gorm:"uniqueIndex:idx_user_note_report" json:"note_id"`
	Reason    string    `gorm:"type:varchar(255)" json:"reason"`                 // 举报理由
	CreatedAt time.Time `json:"created_at"`
}

func (NoteReport) TableName() string { return "note_reports" }