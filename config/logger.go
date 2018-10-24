package config

import "path/filepath"

type LoggerConfig struct {
	LogPath string
}

var ElrondLoggerConfig = LoggerConfig{filepath.Join(DefaultPath(), "logs")}