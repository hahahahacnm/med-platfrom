package user

import (
	"med-platform/internal/common/db"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

// Create 在数据库创建一个新用户
func (r *Repository) Create(user *User) error {
	return db.DB.Create(user).Error
}

// GetByEmail 根据邮箱查找用户 (防止重复注册用)
func (r *Repository) GetByEmail(email string) (*User, error) {
	var user User
	result := db.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}