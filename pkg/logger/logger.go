package logger

import (
	"encoding/json"
	"fmt"
	"time"
)

var LogEnabled = true

type LogEntry = map[string]interface{}

const LogLevelError = "error"
const LogLevelFatal = "fatal"
const LogLevelInfo = "info"

func print(key string, level string, msg interface{}) {
	en := LogEntry{}
	en["key"] = key
	en["level"] = level
	en["message"] = msg
	en["date"] = time.Now().Format(time.RFC3339)

	m, err := json.Marshal(en)
	if err != nil {
		fmt.Println(m)
	}
}

func Error(key string, msg interface{}) {
	if LogEnabled == false {
		return
	}
	print(key, LogLevelError, msg)
}

func Info(key string, msg interface{}) {
	if LogEnabled == false {
		return
	}
	print(key, LogLevelInfo, msg)
}

func Fatal(key string, msg interface{}) {
	if LogEnabled == false {
		return
	}
	print(key, LogLevelFatal, msg)
}