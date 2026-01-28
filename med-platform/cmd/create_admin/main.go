package main

import (
	"fmt"
	"med-platform/internal/common/config"
	"med-platform/internal/common/db"
	"med-platform/internal/user"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	// 1. 初始化配置和数据库
	// 假设 config.Load() 会从 configs/config.yaml 读取
	config.Load()
	db.Init()

	username := "admin"
	password := "admin123"
	role := "admin"

	// 2. 检查用户是否存在
	var u user.User
	result := db.DB.Where("username = ?", username).First(&u)

	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// 创建新用户
			newUser := user.User{
				Username: username,
				Password: string(hashedPwd),
				Role:     role,
				Nickname: "Super Admin",
				Status:   1, // 正常
			}
			if err := db.DB.Create(&newUser).Error; err != nil {
				panic(err)
			}
			fmt.Printf("✅ 已创建管理员账号: %s / %s\n", username, password)
		} else {
			panic(result.Error)
		}
	} else {
		// 更新现有用户 (重置密码为 admin123)
		u.Password = string(hashedPwd)
		u.Role = role
		u.Status = 1
		if err := db.DB.Save(&u).Error; err != nil {
			panic(err)
		}
		fmt.Printf("♻️ 已重置管理员账号: %s / %s\n", username, password)
	}
}
