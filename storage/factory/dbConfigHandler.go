package factory

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/pelletier/go-toml"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
)

const (
	dbConfigFileName         = "config.toml"
	defaultType              = "LvlDBSerial"
	defaultBatchDelaySeconds = 2
	defaultMaxBatchSize      = 100
	defaultMaxOpenFiles      = 10
	defaultUseTmpAsFilePath  = false
	dirPermissions           = 0700
	filePermissions          = 0600
)

var (
	errInvalidConfiguration       = errors.New("invalid configuration")
	errShardedLayoutWithoutConfig = errors.New("sharded persister directory without config file")
)

type dbConfigHandler struct {
	conf config.DBConfig
}

// NewDBConfigHandler will create a new db config handler instance
func NewDBConfigHandler(config config.DBConfig) *dbConfigHandler {
	return &dbConfigHandler{
		conf: config,
	}
}

// GetDBConfig resolves the engine of an existing directory from its config file (legacy dirs:
// from the files it holds) with the main config tuning applied; a new directory takes the main config
func (dh *dbConfigHandler) GetDBConfig(path string) (*config.DBConfig, error) {
	dbConfigFromFile := &config.DBConfig{}
	err := readCorrectConfigurationFromToml(dbConfigFromFile, getPersisterConfigFilePath(path))
	if err == nil {
		dbConfig := dh.withMainConfigTuning(dbConfigFromFile)
		log.Debug("GetDBConfig: loaded db config from toml config file",
			"config path", path,
			"configuration", fmt.Sprintf("%+v", dbConfig),
		)
		return dbConfig, nil
	}

	if checkIfDirIsEmpty(path) {
		log.Debug("GetDBConfig: loaded db config from main config file",
			"configuration", fmt.Sprintf("%+v", dh.conf),
		)
		return &dh.conf, nil
	}

	engineType, err := detectLegacyEngineType(path)
	if err != nil {
		return nil, fmt.Errorf("%w for path %s", err, path)
	}

	dbConfig := &config.DBConfig{
		Type:                  engineType,
		BatchDelaySeconds:     dh.conf.BatchDelaySeconds,
		MaxBatchSize:          dh.conf.MaxBatchSize,
		MaxOpenFiles:          dh.conf.MaxOpenFiles,
		UseTmpAsFilePath:      dh.conf.UseTmpAsFilePath,
		BloomFilterBitsPerKey: dh.conf.BloomFilterBitsPerKey,
		PebbleProfile:         dh.conf.PebbleProfile,
	}

	log.Debug("GetDBConfig: loaded default db config",
		"configuration", fmt.Sprintf("%+v", dbConfig),
	)

	return dbConfig, nil
}

// withMainConfigTuning keeps the persisted identity fields (engine type, on-disk layout) and takes
// the runtime tuning from the main config, so a retune applies to existing directories on restart
func (dh *dbConfigHandler) withMainConfigTuning(persisted *config.DBConfig) *config.DBConfig {
	dbConfig := *persisted
	dbConfig.BatchDelaySeconds = positiveOrFallback(dh.conf.BatchDelaySeconds, persisted.BatchDelaySeconds)
	dbConfig.MaxBatchSize = positiveOrFallback(dh.conf.MaxBatchSize, persisted.MaxBatchSize)
	dbConfig.MaxOpenFiles = positiveOrFallback(dh.conf.MaxOpenFiles, persisted.MaxOpenFiles)
	dbConfig.BloomFilterBitsPerKey = dh.conf.BloomFilterBitsPerKey
	if len(dh.conf.PebbleProfile) > 0 {
		dbConfig.PebbleProfile = dh.conf.PebbleProfile
	}

	return &dbConfig
}

func positiveOrFallback(value int, fallback int) int {
	if value > 0 {
		return value
	}

	return fallback
}

// detectLegacyEngineType classifies a config-less directory by its files, so the goleveldb open (and
// its automatic recovery) never touches a pebble directory; a sharded layout cannot be inferred
func detectLegacyEngineType(path string) (string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}

	engineType := defaultType
	for _, entry := range entries {
		if entry.IsDir() {
			return "", errShardedLayoutWithoutConfig
		}
		if isForeignEngineFile(entry.Name()) {
			engineType = string(storageunit.PebbleDB)
		}
	}

	return engineType, nil
}

// isForeignEngineFile matches file names goleveldb never produces (pebble naming)
func isForeignEngineFile(name string) bool {
	return strings.HasPrefix(name, "OPTIONS-") ||
		strings.HasPrefix(name, "marker.") ||
		strings.HasSuffix(name, ".sst") ||
		strings.HasSuffix(name, ".blob")
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

// SaveDBConfigToFilePath creates the directory if needed and writes the db config atomically;
// it must run before the engine creates any file in the directory
func (dh *dbConfigHandler) SaveDBConfigToFilePath(path string, dbConfig *config.DBConfig) error {
	err := os.MkdirAll(path, dirPermissions)
	if err != nil {
		return err
	}

	data, err := toml.Marshal(dbConfig)
	if err != nil {
		return err
	}

	return writeFileAtomically(getPersisterConfigFilePath(path), data)
}

func writeFileAtomically(filePath string, data []byte) error {
	tmpPath := filePath + ".tmp"
	err := writeAndSyncFile(tmpPath, data)
	if err != nil {
		return err
	}

	err = os.Rename(tmpPath, filePath)
	if err != nil {
		return err
	}

	return syncDir(filepath.Dir(filePath))
}

func writeAndSyncFile(filePath string, data []byte) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, filePermissions)
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	if err != nil {
		_ = file.Close()
		return err
	}

	err = file.Sync()
	if err != nil {
		_ = file.Close()
		return err
	}

	return file.Close()
}

func syncDir(dirPath string) error {
	dir, err := os.Open(dirPath)
	if err != nil {
		return err
	}

	err = dir.Sync()
	if err != nil {
		_ = dir.Close()
		return err
	}

	return dir.Close()
}

// isKnownPersistentDBType is an allow list: an unknown type must fail without leaving a config behind
func isKnownPersistentDBType(dbType string) bool {
	switch storageunit.DBType(dbType) {
	case storageunit.LvlDB, storageunit.LvlDBSerial, storageunit.PebbleDB:
		return true
	default:
		return false
	}
}

// removeConfigIfAlone keeps a directory "new" when the engine creation failed right after the config write
func removeConfigIfAlone(path string) {
	entries, err := os.ReadDir(path)
	if err != nil || len(entries) != 1 || entries[0].Name() != dbConfigFileName {
		return
	}

	_ = os.Remove(getPersisterConfigFilePath(path))
}

func getPersisterConfigFilePath(path string) string {
	return filepath.Join(
		path,
		dbConfigFileName,
	)
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
