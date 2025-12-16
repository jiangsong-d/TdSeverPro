package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
)

// InitLogger 初始化日志
func InitLogger() {
	// 创建日志文件
	logFile, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("无法创建日志文件:", err)
	}
	
	// 多写入器：同时输出到文件和控制台
	infoLogger = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	warnLogger = log.New(os.Stdout, "[WARN] ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(os.Stdout, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
	
	// 同时写入文件
	infoLogger.SetOutput(logFile)
	warnLogger.SetOutput(logFile)
	errorLogger.SetOutput(logFile)
}

// Info 信息日志
func Info(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	if infoLogger != nil {
		infoLogger.Output(2, msg)
	}
	// 同时输出到控制台
	fmt.Printf("[INFO] %s %s\n", time.Now().Format("2006-01-02 15:04:05"), msg)
}

// Warn 警告日志
func Warn(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	if warnLogger != nil {
		warnLogger.Output(2, msg)
	}
	fmt.Printf("[WARN] %s %s\n", time.Now().Format("2006-01-02 15:04:05"), msg)
}

// Error 错误日志
func Error(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	if errorLogger != nil {
		errorLogger.Output(2, msg)
	}
	fmt.Printf("[ERROR] %s %s\n", time.Now().Format("2006-01-02 15:04:05"), msg)
}
