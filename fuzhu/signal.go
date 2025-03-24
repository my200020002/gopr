package fuzhu

import (
	"gopr/fuzhu/logger"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	quitChan    = make(chan os.Signal, 1)
	outputMutex sync.Mutex
)

func handleInterrupt() {
	signal.Notify(quitChan, os.Interrupt, syscall.SIGTERM)
	<-quitChan
	outputMutex.Lock()
	logger.Warn("正在关闭代理服务器...")
	logger.Sync()
	outputMutex.Unlock()
	close(quitChan)
	os.Exit(0)
}
