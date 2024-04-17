package utils

import (
	"encoding/hex"
	"os"
	"path/filepath"
)

func GetSCCode(fileName string) string {
	code, err := os.ReadFile(filepath.Clean(fileName))
	if err != nil {
		panic("Could not get SC code.")
	}

	codeEncoded := hex.EncodeToString(code)
	return codeEncoded
}
