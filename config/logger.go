package config

import "path/filepath"

// LoggerConfig holds the configurable elements for the application logger
type LoggerConfig struct {
	LogPath string
	StackTraceDepth int
}

// ElrondLoggerConfig holds the configuration for the application logger
var ElrondLoggerConfig = LoggerConfig{
	filepath.Join(DefaultPath(), "logs"),
	2,
}
