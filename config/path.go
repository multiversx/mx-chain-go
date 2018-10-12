package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// DefaultPath gives back the path to a default location in user HOME to be used for Elrond application storage
func DefaultPath() string {
	home := os.Getenv("HOME")

	if home != "" {
		switch runtime.GOOS {
		case "windows":
			return filepath.Join(home, "AppData", "Elrond")
		case "linux":
			return filepath.Join(home, ".elrond")
		case "darwin":
			return filepath.Join(home, "Library", "Elrond")
		}
	}
	return ""
}
