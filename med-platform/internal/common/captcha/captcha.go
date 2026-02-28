package captcha

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nfnt/resize"
	"github.com/wenlng/go-captcha/v2/base/option"
	"github.com/wenlng/go-captcha/v2/rotate"
)

// 简单的内存缓存 (你也可以换成 Redis)
var (
	store      sync.Map
	imageCache = make(map[string]image.Image)
	cacheMutex sync.RWMutex
	// 注意：确保这个路径下有你的图片
	imgDir = "uploads/captcha_images" 
)

func Init() {
	// 检查图片目录，没有就创建
	if _, err := os.Stat(imgDir); os.IsNotExist(err) {
		os.MkdirAll(imgDir, os.ModePerm)
	}
	rand.Seed(time.Now().UnixNano())
}

// Generate 生成验证码
func Generate() (string, string, string, error) {
	img, err := getRandomImage()
	if err != nil {
		return "", "", "", err
	}

	builder := rotate.NewBuilder(
		rotate.WithRangeAnglePos([]option.RangeVal{{Min: 30, Max: 330}}),
	)
	builder.SetResources(rotate.WithImages([]image.Image{img}))

	capt := builder.Make()
	data, err := capt.Generate()
	if err != nil {
		return "", "", "", err
	}

	blockData := data.GetData()
	masterBase64, _ := data.GetMasterImage().ToBase64()
	thumbBase64, _ := data.GetThumbImage().ToBase64()

	// 生成唯一 Key
	key := fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Intn(10000))
	
	// 存入缓存 (5分钟有效)
	store.Store(key, blockData.Angle)
	time.AfterFunc(5*time.Minute, func() { store.Delete(key) })

	return key, thumbBase64, masterBase64, nil
}

// Verify 校验
func Verify(key string, answer string) bool {
	val, ok := store.Load(key)
	if !ok {
		return false
	}
	// 验证一次后删除（防重放）
	store.Delete(key)

	correctAngle := val.(int)
	inputAngle, _ := strconv.Atoi(answer)

	// 互补角算法
	diff1 := math.Abs(float64(inputAngle - correctAngle))
	diff2 := math.Abs(float64(inputAngle + correctAngle - 360))

	return diff1 <= 5 || diff2 <= 5
}

// 内部工具：获取图片
func getRandomImage() (image.Image, error) {
	entries, err := ioutil.ReadDir(imgDir)
	if err != nil || len(entries) == 0 {
		return nil, fmt.Errorf("图片目录为空或读取失败: %s", imgDir)
	}
	
	// 简单过滤图片文件
	var files []os.FileInfo
	for _, f := range entries {
		if !f.IsDir() && (strings.HasSuffix(f.Name(), ".jpg") || strings.HasSuffix(f.Name(), ".png")) {
			files = append(files, f)
		}
	}
	if len(files) == 0 { return nil, fmt.Errorf("没有找到图片") }

	file := files[rand.Intn(len(files))]
	
	// 缓存逻辑
	cacheMutex.RLock()
	if img, ok := imageCache[file.Name()]; ok {
		cacheMutex.RUnlock()
		return img, nil
	}
	cacheMutex.RUnlock()

	// 读取并缩放
	f, _ := os.Open(filepath.Join(imgDir, file.Name()))
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil { return nil, err }

	// 统一缩放到 300x300
	resized := resize.Thumbnail(300, 300, img, resize.Lanczos3)
	
	cacheMutex.Lock()
	imageCache[file.Name()] = resized
	cacheMutex.Unlock()

	return resized, nil
}