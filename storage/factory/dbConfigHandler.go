package factory

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
)

const (
	dbConfigFileName = "config.toml"
	defaultType      = "LvlDBSerial"
)

var (
	errInvalidConfiguration = errors.New("invalid configuration")
)

type dbConfigHandler struct {
	dbType              string
	batchDelaySeconds   int
	maxBatchSize        int
	maxOpenFiles        int
	shardIDProviderType string
	numShards           int32
}

// NewDBConfigHandler will create a new db config handler instance
func NewDBConfigHandler(config config.DBConfig) *dbConfigHandler {
	return &dbConfigHandler{
		dbType:              config.Type,
		batchDelaySeconds:   config.BatchDelaySeconds,
		maxBatchSize:        config.MaxBatchSize,
		maxOpenFiles:        config.MaxOpenFiles,
		shardIDProviderType: config.ShardIDProviderType,
		numShards:           config.NumShards,
	}
}

// GetDBConfig will get the db config based on path
func (dh *dbConfigHandler) GetDBConfig(path string) (*config.DBConfig, error) {
	dbConfigFromFile := &config.DBConfig{}
	err := readCorrectConfigurationFromToml(dbConfigFromFile, getPersisterConfigFilePath(path))
	if err == nil {
		log.Debug("GetDBConfig: loaded db config from toml config file",
			"config path", path,
			"configuration", fmt.Sprintf("%+v", dbConfigFromFile),
		)
		return dbConfigFromFile, nil
	}

	empty := checkIfDirIsEmpty(path)
	if !empty {
		dbConfig := &config.DBConfig{
			Type:              defaultType,
			BatchDelaySeconds: dh.batchDelaySeconds,
			MaxBatchSize:      dh.maxBatchSize,
			MaxOpenFiles:      dh.maxOpenFiles,
		}

		log.Debug("GetDBConfig: loaded default db config",
			"configuration", fmt.Sprintf("%+v", dbConfig),
		)

		return dbConfig, nil
	}

	dbConfig := &config.DBConfig{
		Type:                dh.dbType,
		BatchDelaySeconds:   dh.batchDelaySeconds,
		MaxBatchSize:        dh.maxBatchSize,
		MaxOpenFiles:        dh.maxOpenFiles,
		ShardIDProviderType: dh.shardIDProviderType,
		NumShards:           dh.numShards,
	}

	log.Debug("GetDBConfig: loaded db config from main config file",
		"configuration", fmt.Sprintf("%+v", dbConfig),
	)

	return dbConfig, nil
}

func readCorrectConfigurationFromToml(dbConfig *config.DBConfig, filePath string) error {
	err := core.LoadTomlFile(dbConfig, filePath)
	if err != nil {
		return err
	}

	isInvalidConfig := len(dbConfig.Type) == 0 || dbConfig.MaxBatchSize <= 0 || dbConfig.BatchDelaySeconds <= 0 || dbConfig.MaxOpenFiles <= 0
	if isInvalidConfig {
		return errInvalidConfiguration
	}

	return nil
}

// SaveDBConfigToFilePath will save the provided db config to specified path
func (dh *dbConfigHandler) SaveDBConfigToFilePath(path string, dbConfig *config.DBConfig) error {
	pathExists, err := checkIfDirExists(path)
	if err != nil {
		return err
	}
	if !pathExists {
		// in memory db, no files available
		return nil
	}

	configFilePath := getPersisterConfigFilePath(path)

	err = core.SaveTomlFile(dbConfig, configFilePath)
	if err != nil {
		return err
	}

	return nil
}

func getPersisterConfigFilePath(path string) string {
	return filepath.Join(
		path,
		dbConfigFileName,
	)
}

func checkIfDirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func checkIfDirIsEmpty(path string) bool {
	files, err := os.ReadDir(path)
	if err != nil {
		log.Trace("getDBConfig: failed to check if dir is empty", "path", path, "error", err.Error())
		return true
	}

	if len(files) == 0 {
		return true
	}

	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (dh *dbConfigHandler) IsInterfaceNil() bool {
	return dh == nil
}
