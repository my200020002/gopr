package fuzhu

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var (
	// Windows 非法字符: \ / : * ? " < > |
	windowsInvalidChars = regexp.MustCompile(`[\\/:*?"<>|]`)
	// Unix 非法字符: /
	unixInvalidChars = regexp.MustCompile(`/`)
)

// CleanPathReplace 将无效字符替换为下划线
func CleanPathReplace(path string) string {
	if runtime.GOOS == "windows" {
		// Windows 系统
		path = windowsInvalidChars.ReplaceAllString(path, "_")
	} else {
		// Unix 系统
		path = unixInvalidChars.ReplaceAllString(path, "_")
	}
	// 替换空格为下划线
	path = strings.ReplaceAll(path, " ", "_")
	// 清理连续的下划线
	path = regexp.MustCompile(`_+`).ReplaceAllString(path, "_")
	return strings.Trim(path, "_")
}

// CleanPathRemove 清除无效字符
func CleanPathRemove(path string) string {
	if runtime.GOOS == "windows" {
		// Windows 系统
		path = windowsInvalidChars.ReplaceAllString(path, "")
	} else {
		// Unix 系统
		path = unixInvalidChars.ReplaceAllString(path, "")
	}
	// 移除空格
	path = strings.ReplaceAll(path, " ", "")
	return path
}
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}
func GetDesktopPath(fileName string) string {
	return filepath.Join(os.Getenv("USERPROFILE"), "Desktop", fileName)
}
func IsFileExistsWithHash(data []byte, filePath string) bool {
	// 1. 快速检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	// 2. 检查文件大小是否相同
	if fileInfo.Size() != int64(len(data)) {
		return false
	}
	// 3. 如果文件存在且大小相同，读取文件并比对哈希
	existingData, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}
	newHash := CalculateHashFNV(data)
	existingHash := CalculateHashFNV(existingData)
	return newHash == existingHash
}
