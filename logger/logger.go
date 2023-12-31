package logger

import (
	"os"

	log "github.com/sirupsen/logrus"
)

var Log *log.Entry

const (
	debug   = "Debug"
	info    = "Info"
	warning = "Warning"
)

func Init(level string) {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)
	switch level {
	case debug:
		log.SetLevel(log.DebugLevel)
		// log.SetReportCaller(true)
	case info:
		log.SetLevel(log.InfoLevel)
	case warning:
		log.SetLevel(log.WarnLevel)
	}

	Log = log.WithFields(log.Fields{})
}
