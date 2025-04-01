package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var (
	logFile *os.File
	once    sync.Once
)

// 初始化日志，确保只初始化一次
func InitLogger(logDir string) error {
	var err error
	once.Do(func() {
		if err = os.MkdirAll(logDir, 0755); err != nil {
			return
		}
		logPath := filepath.Join(logDir, "task-service.log")
		logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return
		}

		// 设置输出到文件和控制台
		log.SetOutput(os.Stdout)
		if logFile != nil {
			// 同时输出到文件和控制台
			log.SetOutput(io.MultiWriter(os.Stdout, logFile))
		}
		log.SetPrefix("[TASK]")
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	})
	return err
}

// 关闭日志文件
func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}
