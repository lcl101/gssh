package core

import (
	"log"
	"os"
	"path/filepath"
)

type logger struct {
	File  string
	level string
}

//Log 全局log
var Log logger

func init() {
	path, _ := GetExecPath()
	_, fn := filepath.Split(os.Args[0])
	if fn == "" {
		fn = "gssh.log"
	} else {
		fn = fn + ".log"
	}
	logFile, _ := ParsePath(path + fn)
	Log = logger{
		File: logFile,
	}
}

func (logger *logger) write(msg ...interface{}) {
	if _, err := os.Stat(logger.File); err != nil {
		if os.IsNotExist(err) {
			_, err := os.Create(logger.File)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}

	logFile, err := os.OpenFile(logger.File, os.O_RDWR|os.O_APPEND, 0666)
	defer func() {
		if logFile != nil {
			logFile.Close()
		}
	}()
	if err != nil {
		panic(err)
	}

	// 创建一个日志对象
	l := log.New(logFile, logger.level, log.LstdFlags)
	l.Println(msg...)
}

// func (logger *logger) Category(category string) *logger {
// 	logger.category = category
// 	return logger
// }

func (logger *logger) Debug(msg ...interface{}) {
	logger.level = "[D]"
	logger.write(msg...)
}

func (logger *logger) Info(msg ...interface{}) {
	logger.level = "[I]"
	logger.write(msg...)
}

func (logger *logger) Error(msg ...interface{}) {
	logger.level = "[E]"
	logger.write(msg...)
}
