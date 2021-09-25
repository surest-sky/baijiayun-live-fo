package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

var (
	logPath      string
	err          error
	currentTime  string
	logFileName  string
	logErrorName string
)

func init() {
	// 只考虑Mac和Linux系统
	//logPath = os.Getenv("HOME") + "/Desktop/kwh_logs"
	//sysType := runtime.GOOS
	//
	//// Linux 系统
	//if sysType == "linux" {
	//	logPath = "./storage/logs"
	//}

	logPath = "storage/logs"

	// 文件不存在创建文件
	if err = pathExists(logPath); err != nil {
		_ = os.MkdirAll(logPath, os.ModePerm)
	}

	// 初始化文件位置相关
	currentTime = time.Now().Format("2006-01-02")
	logFileName = fmt.Sprintf("%s/%s.log", logPath, currentTime)
	logErrorName = fmt.Sprintf("%s/%s-error.log", logPath, currentTime)

	// 启动写入一下文件
	handle("info", "start", "logger init")

	// 初始化日志相关
	fmt.Println("==== logger init====")
}

// 判断文件夹是否存在
func pathExists(path string) error {
	_, err = os.Stat(path)
	return err
}

// 处理 Error 的情况
// isExit == true 时，将停止运行并输出
func PanicError(err error, source string, isExit bool) {
	if err != nil {
		dispatchNotice(err.Error(), source)
		if isExit == true {
			fmt.Println("错误消息:"+err.Error(), "\n"+"来源: "+source+"\n-----------")
		}
	}
}

// 发送通知
func dispatchNotice(msg string, source string) {
	Error("error : "+source, msg)
}

func handle(level string, title string, content interface{}) {
	var (
		osFile   *os.File
		err      error
		filename string
	)

	message := fmt.Sprintf("[%s] %s : %s \n", level, title, content)

	filename = logFileName
	if strings.ToLower(level) == "error" {
		filename = logErrorName
		fmt.Println(message)
	}

	if osFile, err = os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666); err != nil {
		panic("Logger File Error:" + err.Error())
	}

	log.SetOutput(osFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Printf("[%s] %s : %s \n", level, title, content)
}
