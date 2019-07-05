package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// DefaultPath gives back the path to a default location in the current directory to be used for Elrond application
// storage
func DefaultPath() string {
	workingDirectory, err := os.Getwd()

	if err != nil {
		return ""
	}

	switch runtime.GOOS {
	case "windows":
		return filepath.Join(workingDirectory, "Elrond")
	case "linux":
		return filepath.Join(workingDirectory, ".elrond")
	case "darwin":
		return filepath.Join(workingDirectory, "Elrond")
	}

	return "elrond"
}
