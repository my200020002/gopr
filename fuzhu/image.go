package fuzhu

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	// 用于缓存已保存文件的哈希值
	imageHashes = make(map[uint64]bool)
	hashMutex   sync.RWMutex
)

func IsImageResponse(resp *http.Response) bool {
	contentType := resp.Header.Get("Content-Type")
	return strings.HasPrefix(contentType, "image/") ||
		strings.HasSuffix(resp.Request.URL.Path, ".jpg") ||
		strings.HasSuffix(resp.Request.URL.Path, ".jpeg") ||
		strings.HasSuffix(resp.Request.URL.Path, ".png") ||
		strings.HasSuffix(resp.Request.URL.Path, ".gif") ||
		strings.HasSuffix(resp.Request.URL.Path, ".webp")
}

func SaveImage(data []byte, url string) error {
	// 1. 先检查内存缓存中是否存在相同哈希
	hash := CalculateHashFNV(data)
	hashMutex.RLock()
	exists := imageHashes[hash]
	hashMutex.RUnlock()
	if exists {
		return nil
	}

	// 2. 处理文件名
	fileName := path.Base(url)
	fileName = CleanPathReplace(fileName)
	if fileName == "" || fileName == "." || !hasImageExtension(fileName) {
		ext := getImageExtension(url)
		fileName = fmt.Sprintf("image_%d%s", time.Now().UnixNano(), ext)
	}

	// 3. 获取完整路径
	filePath := filepath.Join(os.Getenv("USERPROFILE"), "Desktop", fileName)

	// 4. 检查文件是否存在及大小
	fileInfo, err := os.Stat(filePath)
	if err == nil && fileInfo.Size() == int64(len(data)) {
		// 文件存在且大小相同时才读取进行比对
		existingData, err := os.ReadFile(filePath)
		if err == nil {
			existingHash := CalculateHashFNV(existingData)
			if existingHash == hash {
				// 内容相同，更新缓存并返回
				hashMutex.Lock()
				imageHashes[hash] = true
				hashMutex.Unlock()
				return nil
			}
		}
	}

	// 5. 写入文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return err
	}

	// 6. 更新哈希缓存
	hashMutex.Lock()
	imageHashes[hash] = true
	hashMutex.Unlock()

	return nil
}

// 新增：检查文件是否具有图片扩展名
func hasImageExtension(fileName string) bool {
	ext := strings.ToLower(path.Ext(fileName))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" ||
		ext == ".gif" || ext == ".webp"
}

// 新增：从URL或Content-Type获取适当的图片扩展名
func getImageExtension(url string) string {
	ext := path.Ext(url)
	if ext != "" {
		return ext
	}
	return ".jpg" // 默认扩展名
}
