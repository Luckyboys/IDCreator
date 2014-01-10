package Common

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	NOTICE  = iota
	INFO    = iota
	WARNING = iota
	ERROR   = iota
)

type Log struct {
	writer *log.Logger
	isInit bool
}

//TODO 日志分文件写入
var instanceLog *Log = new(Log)

func GetLogger() *Log {
	if !instanceLog.isInit {
		instanceLog.isInit = true
		logfile, err := os.OpenFile(GetConfigInstance().Get("logpath", "/tmp/idcreator.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)

		if err != nil {
			fmt.Printf("%s\r\n", err.Error())
			os.Exit(-1)
		}

		instanceLog.writer = log.New(logfile, "", 0)
	}

	return instanceLog
}

func (this *Log) CheckError(err error, level int) bool {
	if err != nil {
		go this.WriteLog(err.Error(), level)
		return true
	}

	return false
}

func (this *Log) WriteLog(message string, level int) {

	configLevel, _ := strconv.Atoi(GetConfigInstance().Get("log", "0"))

	if configLevel > level {
		return
	}

	var levelString string

	switch level {
	case ERROR:

		levelString = "ERROR"

	case WARNING:

		levelString = "WARNING"

	case INFO:

		levelString = "INFO"

	case NOTICE:

		levelString = "NOTICE"
	}

	var now = time.Now()
	this.writer.Printf("[%d-%02d-%02d %02d:%02d:%02d][%s]:\t%s", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), levelString, message)
}
