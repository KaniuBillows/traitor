package logger

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	mu      sync.Mutex
	logger  *log.Logger
	_logger *log.Logger
)

// prefix
const (
	DEBUG = "DEBUG"
	INFO  = "INFO"
	WARN  = "WARN"
	ERROR = "ERROR"
	FATAL = "FATAL"
)

func init() {
	dir, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	var path = fmt.Sprintf("%s/.traitor/debug", dir)
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		panic(err)
	}
	var file = path + "/traitorLog.txt"

	logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		panic(err)
	}
	logger = log.New(logFile, "", log.LstdFlags|log.Lshortfile|log.LUTC)
	_logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile|log.LUTC)
}
func Debug(v ...any) {
	_log(DEBUG, v)
}
func Error(v ...any) {
	_log(ERROR, v)
}
func Info(v ...any) {
	_log(INFO, v)
}
func Warn(v ...any) {
	_log(WARN, v)
}
func Fatal(v ...any) {
	_log(FATAL, v)
}
func _log(prefix string, v any) {
	mu.Lock()
	defer mu.Unlock()
	setPrefix(prefix)
	logger.Println(v)
	_logger.Println(v)
}
func setPrefix(logType string) {
	_, file, line, ok := runtime.Caller(2)
	var logPrefix string
	if ok {
		logPrefix = fmt.Sprintf("[%s][%s:%d]", logType, filepath.Base(file), line)
	} else {
		logPrefix = fmt.Sprintf("[%s]", logType)
	}
	logger.SetPrefix(logPrefix)
	_logger.SetPrefix(logPrefix)
}
