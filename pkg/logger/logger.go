package logger

import (
	"encoding/json"
	"fmt"
	"time"
)

var LogEnabled = true

type LogEntry struct {
	Level string `json:"level"`
	Date string `json:"date"`
	Key string `json:"key"`
	Message interface{} `json:"message"`
}

const LogLevelError = "error"
const LogLevelFatal = "fatal"
const LogLevelInfo = "info"

func print(key string, level string, msg interface{}) {
	en := LogEntry{}
	en.Key= key
	en.Level = level
	en.Message = msg
	en.Date = time.Now().Format(time.RFC3339)

	m, err := json.Marshal(en)
	if err == nil {
		fmt.Println(string(m))
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