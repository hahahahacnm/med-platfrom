package uploader

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 配置常量
const (
	MaxAvatarSize    = 2 * 1024 * 1024 // 2MB
	MaxNoteImageSize = 8 * 1024 * 1024 // 8MB
	UploadRootDir    = "./uploads"
	TempDir          = "temp" // 临时池
)

// AllowedExtensions 允许的图片格式
var AllowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

// CalculateFileHash 计算文件内容的 SHA256 哈希值
func CalculateFileHash(fileHeader *multipart.FileHeader) (string, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, src); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// SaveImage 处理普通图片上传 (不计算哈希，直接保存，用于头像等简单场景)
func SaveImage(c *gin.Context, fileKey string, subDir string, maxSize int64) (string, error) {
	file, err := c.FormFile(fileKey)
	if err != nil {
		return "", errors.New("无法获取上传文件")
	}

	if file.Size > maxSize {
		return "", fmt.Errorf("文件大小超过限制 (最大 %.2f MB)", float64(maxSize)/1024/1024)
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !AllowedExtensions[ext] {
		return "", errors.New("不支持的文件格式")
	}

	fileName := fmt.Sprintf("%d_%s%s", time.Now().Unix(), uuid.New().String(), ext)
	
	// 确保目录存在
	finalDir := fmt.Sprintf("%s/%s", UploadRootDir, subDir)
	if subDir == "" {
		finalDir = UploadRootDir
	}
	if _, err := os.Stat(finalDir); os.IsNotExist(err) {
		os.MkdirAll(finalDir, 0755)
	}

	finalPath := fmt.Sprintf("%s/%s", finalDir, fileName)
	
	if err := c.SaveUploadedFile(file, finalPath); err != nil {
		return "", errors.New("文件保存失败")
	}

	return finalPath[1:], nil
}

// SaveImageWithHash 保存图片（哈希去重 + 临时目录策略）
// 注意：这里为了兼容，默认检查 notes 目录的去重，这对 Feedback 也无伤大雅
func SaveImageWithHash(c *gin.Context, fileKey string, maxSize int64) (string, error) {
	file, err := c.FormFile(fileKey)
	if err != nil {
		return "", errors.New("无法获取上传文件")
	}

	// 1. 大小检查
	if file.Size > maxSize {
		return "", fmt.Errorf("文件大小超过限制 (最大 %.2f MB)", float64(maxSize)/1024/1024)
	}

	// 2. 后缀检查
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !AllowedExtensions[ext] {
		return "", errors.New("不支持的文件格式")
	}

	// 3. 🔥 核心：计算哈希，生成唯一文件名
	hashName, err := CalculateFileHash(file)
	if err != nil {
		return "", errors.New("文件解析失败")
	}
	finalFileName := hashName + ext

	// 4. 检查文件是否已经在正式目录 (uploads/notes) 存在？(秒传检查)
	// (即便我们稍后要把Feedback图片存到 feedback 目录，这里检查一下 notes 也没坏处，能省则省)
	finalDestDir := fmt.Sprintf("%s/notes", UploadRootDir)
	finalDestPath := fmt.Sprintf("%s/%s", finalDestDir, finalFileName)
	if _, err := os.Stat(finalDestPath); err == nil {
		return finalDestPath[1:], nil // 秒传
	}

	// 5. 检查文件是否在临时目录 (uploads/temp) 存在？
	tempDestDir := fmt.Sprintf("%s/%s", UploadRootDir, TempDir)
	tempDestPath := fmt.Sprintf("%s/%s", tempDestDir, finalFileName)
	
	if _, err := os.Stat(tempDestDir); os.IsNotExist(err) {
		os.MkdirAll(tempDestDir, 0755)
	}

	if _, err := os.Stat(tempDestPath); err == nil {
		return tempDestPath[1:], nil
	}

	// 6. 保存到临时目录
	if err := c.SaveUploadedFile(file, tempDestPath); err != nil {
		return "", errors.New("文件保存失败")
	}

	return tempDestPath[1:], nil 
}

// 🔥 ConfirmImages 固化图片：将图片从 temp 移动到【指定】的目标目录
// subDir: 例如 "notes", "feedback"
func ConfirmImages(imagePaths []string, subDir string) []string {
	var finalPaths []string
	
	// 确保目标目录存在 (例如 uploads/feedback)
	targetDir := fmt.Sprintf("%s/%s", UploadRootDir, subDir)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		os.MkdirAll(targetDir, 0755)
	}

	for _, path := range imagePaths {
		// 只有在 temp 目录下的才需要移动
		if strings.Contains(path, "/uploads/temp/") {
			// 转换路径: /uploads/temp/abc.jpg -> ./uploads/temp/abc.jpg
			srcPath := "." + path
			fileName := filepath.Base(srcPath)
			destPath := fmt.Sprintf("%s/%s", targetDir, fileName)

			// 移动文件 (如果目标已存在则覆盖/忽略，因为哈希一致内容一致)
			err := os.Rename(srcPath, destPath)
			if err == nil {
				// 移动成功，返回新路径
				newUrl := fmt.Sprintf("/uploads/%s/%s", subDir, fileName)
				finalPaths = append(finalPaths, newUrl)
			} else {
				// 如果移动失败（比如文件已经被移走了，或者目标已存在）
				if _, err := os.Stat(destPath); err == nil {
					// 目标已存在，视为成功
					newUrl := fmt.Sprintf("/uploads/%s/%s", subDir, fileName)
					finalPaths = append(finalPaths, newUrl)
				} else {
					// 真的丢了，保留原路径防止报错，或者返回空
					finalPaths = append(finalPaths, path) 
				}
			}
		} else {
			// 已经是永久路径，直接保留
			finalPaths = append(finalPaths, path)
		}
	}
	return finalPaths
}