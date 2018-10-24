package logger

import (
	"elrond/elrond-go-sandbox/config"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
)

func init() {
	os.MkdirAll(config.ElrondLoggerConfig.LogPath, os.ModePerm)
	file, err := currentLogFile()
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

func currentLogFile() (*os.File, error) {
	return os.OpenFile(
		filepath.Join(config.ElrondLoggerConfig.LogPath, time.Now().Format("01-02-2006") + ".log"),
		os.O_APPEND|os.O_WRONLY,
		0666)
}
