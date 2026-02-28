package sysconfig

import "gorm.io/gorm"

// SysConfig 平台动态参数表
type SysConfig struct {
	gorm.Model
	Key         string `gorm:"type:varchar(64);uniqueIndex;not null" json:"key"`
	Value       string `gorm:"type:text;not null" json:"value"`
	Description string `gorm:"type:varchar(255)" json:"description"`
}