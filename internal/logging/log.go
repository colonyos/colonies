package logging

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func Log() *logrus.Logger {
	return log
}

func Disable() {
	logrus.SetOutput(ioutil.Discard)
}

func EnableDebug() {
	log.Level = logrus.DebugLevel
}

func DisableDebug() {
	log.Level = logrus.ErrorLevel
}
