package logger

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

var LogEnabled = true

const LogLevelError = "error"
const LogLevelFatal = "fatal"
const LogLevelInfo = "info"

const LogModeFile = "file"
const LogModeStdErr = "stderr"

const DefaultDirectory = "./logs/"

var messageTpl = "[%level] %key: %msg %date\n"

var logChan = make(chan string, 10)

var initialized = false

var logFile *os.File
// if logMode is "file", there fileName should be passed,
// else, it must be passed nil
func Initialize(logMode string, fileName string) error {
	initialized = true
	var err error
	if logFile == nil {
		dir := path.Dir(fileName)
		err = os.Mkdir(dir, os.FileMode(0775))
		if err != nil  && !strings.Contains(err.Error(), "exists") {
			return err
		}
		logFile, err = os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.FileMode(0644))
		if err != nil {
			return err
		}
	}
	go func() {
		sendLogMessage(logMode, logFile)
	}()
	return nil
}

func sendLogMessage(logMode string, fd *os.File) {
	if logMode == LogModeFile {
		defer fd.Close()
		for msg := range logChan {
			if fd  == nil {
				panic("log mode is set to file, but no existing file specified")
			}
			_, err := fd.WriteString(msg)
			if err != nil {
				log.Println("failed to insert log into file, "+err.Error())
			}
		}
	} else if logMode == LogModeFile {
		for msg := range logChan {
			fmt.Println(msg)
		}
	}
}

func print(key string, level string, msg interface{}) {
	if initialized == false {
		panic("logger is not initialized")
	}
	var prs = strings.Replace(messageTpl, "%level", level, 1)
	prs = strings.Replace(prs, "%key", key, 1)
	prs = strings.Replace(prs, "%msg", fmt.Sprintf("%v", msg), 1)
	prs = strings.Replace(prs, "%date", time.Now().Format(time.RFC3339), 1)
	logChan <- prs
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

func InfoOut(key string, msg interface{}) {
	if LogEnabled == false {
		return
	}
	if initialized == false {
		panic("logger is not initialized")
	}
	var prs = strings.Replace(messageTpl, "%level", LogLevelInfo, 1)
	prs = strings.Replace(prs, "%key", key, 1)
	prs = strings.Replace(prs, "%msg", fmt.Sprintf("%v", msg), 1)
	prs = strings.Replace(prs, "%date", time.Now().Format(time.RFC3339), 1)
	fmt.Print(prs)
}


func Fatal(key string, msg interface{}) {
	if LogEnabled == false {
		return
	}
	print(key, LogLevelFatal, msg)
}
