package main

import (
	"fmt"
	"log"
	"os"
)

var DebugMode = false

// infoLogger 输出到 stdout
var infoLogger = log.New(os.Stdout, "[INFO] ", log.LstdFlags)

func LogInfo(msgs ...interface{}) {
	infoLogger.Println(msgs...)
}

var debugLogger = log.New(os.Stdout, "[DEBUG] ", log.LstdFlags)

func LogDebug(msgs ...interface{}) {
	if DebugMode {
		debugLogger.Println(msgs...)
	}
}

// 创建自定义 Logger，启用行号和文件名输出
var errorLogger = log.New(
	os.Stderr,
	"[ERROR] ",
	log.LstdFlags|log.Lshortfile, // 包含日期、时间、短文件名和行号
)

func LogError(msgs ...interface{}) {
	// 调用 Output 方法，calldepth=2 表示跳过 Logger 自身的调用栈
	errorLogger.Output(2, fmt.Sprintln(msgs...))
}
