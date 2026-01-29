package question

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Category 分类表 (目录树)
type Category struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	ParentID  *uint          `gorm:"index" json:"parent_id"`
	Level     int            `gorm:"default:1" json:"level"`
	SortOrder int            `gorm:"default:999" json:"sort_order"`
	FullPath  string         `gorm:"type:text;index" json:"full_path"`
	IsDirty   bool           `gorm:"default:false" json:"is_dirty"`
	Source    string         `gorm:"index;size:100;not null;default:''"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`

	// 🔥🔥🔥 核心修复：补上 IsLeaf 字段 🔥🔥🔥
	IsLeaf    bool           `gorm:"-" json:"is_leaf"`
}

func (Category) TableName() string {
	return "categories"
}

// UserRecordDTO (保留)
type UserRecordDTO struct {
	Choice    string `json:"choice"`
	IsCorrect bool   `json:"is_correct"`
}

// Question 题目表
type Question struct {
	ID uint `gorm:"primaryKey" json:"id"`

	CategoryPath string `gorm:"type:varchar(255);index" json:"category_path,omitempty"`
	Category     string `gorm:"type:varchar(100);index" json:"category,omitempty"`

	// 🔥🔥🔥 新增：必须加这个字段，否则统计代码无法通过 ID 关联科目 🔥🔥🔥
	CategoryID   uint   `gorm:"index;default:0" json:"category_id"`

	Source       string `gorm:"type:varchar(100);index" json:"source,omitempty"`
	Subject      string `gorm:"type:varchar(50);index" json:"subject,omitempty"`
	Chapter      string `gorm:"type:varchar(50)" json:"chapter,omitempty"`

	Type     string         `gorm:"type:varchar(20);index" json:"type"`
	Stem     string         `gorm:"type:text;not null" json:"stem"`
	Material string         `gorm:"type:text" json:"material,omitempty"`
	Options  datatypes.JSON `gorm:"type:jsonb" json:"options,omitempty"`
	Correct  string         `gorm:"type:text" json:"correct,omitempty"`
	Analysis string         `gorm:"type:text" json:"analysis,omitempty"`

	ParentID *uint      `gorm:"index" json:"-"`
	Parent   *Question  `gorm:"foreignKey:ParentID" json:"-"`
	Children []Question `gorm:"foreignKey:ParentID" json:"children,omitempty"`

	Difficulty     string  `gorm:"type:varchar(10)" json:"difficulty,omitempty"`
	DiffValue      float64 `gorm:"type:decimal(3,2)" json:"diff_value,omitempty"`
	CognitiveLevel string  `gorm:"type:varchar(20)" json:"cognitive_level,omitempty"`
	Syllabus       string  `gorm:"size:50" json:"syllabus"`

	UserRecord interface{} `gorm:"-" json:"user_record,omitempty"`
	
	// 🔥 确保这个字段也在
	NoteCount  int64       `gorm:"-" json:"note_count"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (Question) TableName() string {
	return "questions"
}

// ---------------------------------------------------------
// 🔥🔥🔥 新增统计表 (为了实现冷热分离 + 永久记录) 🔥🔥🔥
// ---------------------------------------------------------

// UserDailyStat 用户每日刷题统计 (热数据)
// 作用：记录近一年的每日做题量，用于热力图、连续打卡。
// 特点：哪怕题目重置了，这里的数据也不会删，保证"苦劳"被记录。
type UserDailyStat struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"index:idx_user_date,unique" json:"user_id"` // 联合唯一索引
	DateStr   string    `gorm:"index:idx_user_date,unique;type:char(10)" json:"date_str"` // 格式 "2024-01-28"
	Count     int       `json:"count"` // 当天累计做题数 (只增不减)
	UpdatedAt time.Time `json:"-"`
}

func (UserDailyStat) TableName() string {
	return "user_daily_stats"
}

// UserArchivedStat 用户历史归档统计 (冷数据)
// 作用：存储超过365天的老数据总和，保证"总刷题数"不丢失。
// 特点：每个用户永远只有一行数据。
type UserArchivedStat struct {
	UserID     uint      `gorm:"primaryKey" json:"user_id"` // UserID 作为主键
	TotalCount int64     `json:"total_count"`               // 历史陈年旧账的总和
	UpdatedAt  time.Time `json:"-"`
}

func (UserArchivedStat) TableName() string {
	return "user_archived_stats"
}