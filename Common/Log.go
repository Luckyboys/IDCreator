package Common

import (
	"log"
)

const (
	NOTICE  = iota
	INFO    = iota
	WARNING = iota
	ERROR   = iota
)

func CheckError(err error, level int) bool {
	if err != nil {
		go WriteLog(err.Error(), level)
		return true
	}

	return false
}

func WriteLog(message string, level int) {
	var levelString string

	switch level {
	case ERROR:

		levelString = "ERROR"

	case WARNING:

		levelString = "WARNING"

	case INFO:

		levelString = "INFO"

	case NOTICE:
		return
		levelString = "NOTICE"
	}

	log.Printf("[%s]:\t%s", levelString, message)
}
