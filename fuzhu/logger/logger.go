package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/pterm/pterm"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	defaultLogger *LoggerWrapper
	fileLogger    *log.Logger
)

type LoggerWrapper struct {
	*pterm.Logger
}

func init() {
	logDir := filepath.Join("logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(err)
	}
	fileWriter := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "app.log"),
		MaxSize:    100,   // 每个日志文件最大尺寸，单位MB
		MaxBackups: 7,     // 保留7个备份
		MaxAge:     7,     // 保留7天
		Compress:   false, // 压缩旧文件
	}

	fileLogger = log.New(fileWriter, "", 0)
	defaultLogger = &LoggerWrapper{
		Logger: pterm.DefaultLogger.
			WithLevel(pterm.LogLevelTrace).
			WithTime(true).
			WithCaller(false),
	}
}
func getCaller() string {
	_, file, line, ok := runtime.Caller(2) // 跳过包装函数
	if !ok {
		return "unknown"
	}
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}
func getFormattedTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// Sync 确保所有日志都写入磁盘
func Sync() {
	if fileLogger != nil {
		if w, ok := fileLogger.Writer().(*lumberjack.Logger); ok {
			w.Rotate()
		}
	}
}

// 导出全局方法
func Info(v ...interface{}) {
	msg := fmt.Sprint(v...)
	fileLogger.Printf("%s %s INF: %s", getFormattedTime(), getCaller(), msg)
	defaultLogger.Logger.WithCallerOffset(1).Info(msg)
}
func Debug(v ...interface{}) {
	msg := fmt.Sprint(v...)
	// fileLogger.Printf("%s %s DBG: %s", getFormattedTime(), getCaller(), msg)
	defaultLogger.Logger.WithCallerOffset(1).Debug(msg)
}
func Warn(v ...interface{}) {
	msg := fmt.Sprint(v...)
	fileLogger.Printf("%s %s WRN: %s", getFormattedTime(), getCaller(), msg)
	defaultLogger.Logger.WithCallerOffset(1).Warn(msg)
}
func Error(v ...interface{}) {
	msg := fmt.Sprint(v...)
	fileLogger.Printf("%s %s ERR: %s", getFormattedTime(), getCaller(), msg)
	defaultLogger.Logger.WithCallerOffset(1).Error(msg)
}
func Fatal(v ...interface{}) {
	msg := fmt.Sprint(v...)
	fileLogger.Printf("%s %s FTL: %s", getFormattedTime(), getCaller(), msg)
	defaultLogger.Logger.WithCallerOffset(1).Fatal(msg)
}
func Print(v ...interface{}) {
	msg := fmt.Sprint(v...)
	fileLogger.Printf("%s %s INF: %s", getFormattedTime(), getCaller(), msg)
	defaultLogger.Logger.WithCallerOffset(1).Info(msg)
}

func Infof(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	fileLogger.Printf("%s %s INF: %s", getFormattedTime(), getCaller(), msg)
	defaultLogger.Logger.WithCallerOffset(1).Info(msg)
}
func Debugf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	// fileLogger.Printf("%s %s DBG: %s", getFormattedTime(), getCaller(), msg)
	defaultLogger.Logger.WithCallerOffset(1).Debug(msg)
}
func Warnf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	fileLogger.Printf("%s %s WRN: %s", getFormattedTime(), getCaller(), msg)
	defaultLogger.Logger.WithCallerOffset(1).Warn(msg)
}
func Errorf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	fileLogger.Printf("%s %s ERR: %s", getFormattedTime(), getCaller(), msg)
	defaultLogger.Logger.WithCallerOffset(1).Error(msg)
}
func Fatalf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	fileLogger.Printf("%s %s FTL: %s", getFormattedTime(), getCaller(), msg)
	defaultLogger.Logger.WithCallerOffset(1).Fatal(msg)
}
func Printf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	fileLogger.Printf("%s %s INF: %s", getFormattedTime(), getCaller(), msg)
	defaultLogger.Logger.WithCallerOffset(1).Info(msg)
}
