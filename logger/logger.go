package logger

import (
	"github.com/sirupsen/logrus"
	"os"
	"cronnest/configure"
)

var Logger = logrus.New()

func init() {
	Logger.Out = os.Stdout
	file, err := os.OpenFile(configure.Log["cronnest"], os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		Logger.Out = file
	}
}