package logger

import (
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"sync"
)

var (
	logFile *os.File
	once    sync.Once
	Logger  *logrus.Logger
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

		log := logrus.New()
		log.SetLevel(logrus.DebugLevel)
		// 设置输出到文件和控制台
		log.SetOutput(os.Stdout)
		if logFile != nil {
			// 同时输出到文件和控制台
			log.SetOutput(io.MultiWriter(os.Stdout, logFile))
		}
		// 设置日志为json输出
		log.SetFormatter(&logrus.JSONFormatter{})
		Logger = log
	})
	return err
}

// 关闭日志文件
func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}
