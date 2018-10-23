package logger

import (
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

func init() {
	file, err := os.OpenFile("../logs/logrus.log", os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		mw := io.MultiWriter(os.Stdout, file)
		log.SetOutput(mw)
	} else {
		fmt.Println(err)
		log.SetOutput(os.Stdout)
	}
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.WarnLevel)
}

func InvalidArgument(name string, value interface{}) {
	log.WithFields(log.Fields{
		"type": "invalid_argument",
		"argument": name,
		"value": value,
	}).Warn("Invalid argument provided")
}

