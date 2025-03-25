package fuzhu

import (
	"bufio"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
)

type FileManager struct {
	encodings map[string]encoding.Encoding
}

func NewFileManager() *FileManager {
	return &FileManager{
		encodings: map[string]encoding.Encoding{
			"utf8":    unicode.UTF8,
			"utf16":   unicode.UTF16(unicode.LittleEndian, unicode.UseBOM),
			"gbk":     simplifiedchinese.GBK,
			"gb18030": simplifiedchinese.GB18030,
			"big5":    traditionalchinese.Big5,
		},
	}
}
func (fm *FileManager) ReadAllBytes(path string) ([]byte, error) {
	return os.ReadFile(path)
}
func (fm *FileManager) ReadAllLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 预分配切片容量
	lines := make([]string, 0, 1000)

	scanner := bufio.NewScanner(file)
	// 设置更大的缓冲区
	buf := make([]byte, 0, 128*1024) // 128KB 的缓冲区
	scanner.Buffer(buf, 1024*1024)   // 保持最大token大小为1MB

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func (fm *FileManager) ReadAllLinesWithEncoding(path string, encodingName string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	enc, ok := fm.encodings[encodingName]
	if !ok {
		return fm.ReadAllLines(path)
	}

	decoder := enc.NewDecoder()
	reader := decoder.Reader(file)

	// 预分配切片容量
	lines := make([]string, 0, 1000)

	scanner := bufio.NewScanner(reader)
	// 设置更大的缓冲区
	buf := make([]byte, 0, 128*1024) // 128KB 的缓冲区
	scanner.Buffer(buf, 1024*1024)   // 保持最大token大小为1MB

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
func (fm *FileManager) ReadAllText(path string) (string, error) {
	bytes, err := fm.ReadAllBytes(path)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
func (fm *FileManager) ReadAllTextWithEncoding(path string, encodingName string) (string, error) {
	bytes, err := fm.ReadAllBytes(path)
	if err != nil {
		return "", err
	}

	enc, ok := fm.encodings[encodingName]
	if !ok {
		return string(bytes), nil // 如果找不到指定编码，默认使用UTF-8
	}

	decoder := enc.NewDecoder()
	result, err := decoder.Bytes(bytes)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

func (fm *FileManager) AppendAllBytes(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	return err
}

func (fm *FileManager) AppendAllLines(path string, lines []string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return err
		}
	}
	return writer.Flush()
}

func (fm *FileManager) AppendAllLinesWithEncoding(path string, lines []string, encodingName string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	enc, ok := fm.encodings[encodingName]
	if !ok {
		return fm.AppendAllLines(path, lines)
	}

	encoder := enc.NewEncoder()
	writer := bufio.NewWriter(encoder.Writer(file))
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return err
		}
	}
	return writer.Flush()
}

func (fm *FileManager) AppendAllText(path string, text string) error {
	return fm.AppendAllBytes(path, []byte(text))
}

func (fm *FileManager) AppendAllTextWithEncoding(path string, text string, encodingName string) error {
	enc, ok := fm.encodings[encodingName]
	if !ok {
		return fm.AppendAllText(path, text)
	}

	encoder := enc.NewEncoder()
	data, err := encoder.Bytes([]byte(text))
	if err != nil {
		return err
	}
	return fm.AppendAllBytes(path, data)
}

func (fm *FileManager) WriteAllBytes(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

func (fm *FileManager) WriteAllLines(path string, lines []string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return err
		}
	}
	return writer.Flush()
}

func (fm *FileManager) WriteAllLinesWithEncoding(path string, lines []string, encodingName string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	enc, ok := fm.encodings[encodingName]
	if !ok {
		return fm.WriteAllLines(path, lines)
	}

	encoder := enc.NewEncoder()
	writer := bufio.NewWriter(encoder.Writer(file))
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return err
		}
	}
	return writer.Flush()
}

func (fm *FileManager) WriteAllText(path string, text string) error {
	return fm.WriteAllBytes(path, []byte(text))
}

func (fm *FileManager) WriteAllTextWithEncoding(path string, text string, encodingName string) error {
	enc, ok := fm.encodings[encodingName]
	if !ok {
		return fm.WriteAllText(path, text)
	}

	encoder := enc.NewEncoder()
	data, err := encoder.Bytes([]byte(text))
	if err != nil {
		return err
	}
	return fm.WriteAllBytes(path, data)
}
func (fm *FileManager) Copy(srcPath string, destPath string) error {
	// 检查源文件是否存在
	if !fm.Exists(srcPath) {
		return os.ErrNotExist
	}

	// 确保目标目录存在
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	// 打开源文件
	source, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer source.Close()

	// 创建目标文件
	destination, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destination.Close()

	// 使用缓冲区复制文件
	buf := make([]byte, 128*1024) // 32KB 缓冲区
	_, err = io.CopyBuffer(destination, source, buf)
	return err
}

func (fm *FileManager) Delete(path string) bool {
	err := os.Remove(path)
	return err == nil
}

func (fm *FileManager) Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
func (fm *FileManager) GetSupportedEncodings() []string {
	encodings := make([]string, 0, len(fm.encodings))
	for k := range fm.encodings {
		encodings = append(encodings, k)
	}
	return encodings
}
