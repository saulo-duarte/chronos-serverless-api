package config

import (
	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func Init() {
	Log = logrus.New()
	Log.SetFormatter(&logrus.JSONFormatter{})
	Log.SetLevel(logrus.InfoLevel)
}
