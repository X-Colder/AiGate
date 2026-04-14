package logger

import (
	"log"
	"os"
	"strings"
)

var (
	level  int
	infoL  *log.Logger
	warnL  *log.Logger
	errorL *log.Logger
	debugL *log.Logger
)

const (
	LevelDebug = iota
	LevelInfo
	LevelWarn
	LevelError
)

func Init(levelStr string) {
	switch strings.ToLower(levelStr) {
	case "debug":
		level = LevelDebug
	case "info":
		level = LevelInfo
	case "warn":
		level = LevelWarn
	case "error":
		level = LevelError
	default:
		level = LevelInfo
	}

	debugL = log.New(os.Stdout, "[DEBUG] ", log.LstdFlags|log.Lshortfile)
	infoL = log.New(os.Stdout, "[INFO]  ", log.LstdFlags|log.Lshortfile)
	warnL = log.New(os.Stdout, "[WARN]  ", log.LstdFlags|log.Lshortfile)
	errorL = log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lshortfile)
}

func Debugf(format string, v ...interface{}) {
	if level <= LevelDebug {
		debugL.Printf(format, v...)
	}
}

func Infof(format string, v ...interface{}) {
	if level <= LevelInfo {
		infoL.Printf(format, v...)
	}
}

func Warnf(format string, v ...interface{}) {
	if level <= LevelWarn {
		warnL.Printf(format, v...)
	}
}

func Errorf(format string, v ...interface{}) {
	if level <= LevelError {
		errorL.Printf(format, v...)
	}
}
