package processingOnlyNode

import (
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml"
)

// LoadConfigFromFile will try to load the config from the specified file
func LoadConfigFromFile(filename string, config interface{}) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	err = toml.Unmarshal(data, config)

	return err
}

// GetLatestGasScheduleFilename will parse the provided path and get the latest gas schedule filename
func GetLatestGasScheduleFilename(directory string) (string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return "", err
	}

	extension := ".toml"
	versionMarker := "V"

	highestVersion := 0
	filename := ""
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		splt := strings.Split(name, versionMarker)
		if len(splt) != 2 {
			continue
		}

		versionAsString := splt[1][:len(splt[1])-len(extension)]
		number, errConversion := strconv.Atoi(versionAsString)
		if errConversion != nil {
			continue
		}

		if number > highestVersion {
			highestVersion = number
			filename = name
		}
	}

	return path.Join(directory, filename), nil
}
