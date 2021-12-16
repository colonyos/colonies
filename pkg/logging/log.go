package logging

import "github.com/sirupsen/logrus"

var log = logrus.New()

func Log() *logrus.Logger {
	return log
}
