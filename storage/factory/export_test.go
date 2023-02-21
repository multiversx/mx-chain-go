package factory

import "github.com/multiversx/mx-chain-go/config"

// GetDBConfig -
func (pf *PersisterFactory) GetDBConfig(path string) (*config.DBConfig, error) {
	return pf.getDBConfig(path)
}

// CreatePersisterConfigFile -
func CreatePersisterConfigFile(path string, dbConfig *config.DBConfig) error {
	return createPersisterConfigFile(path, dbConfig)
}

// GetPersisterConfigFilePath -
func GetPersisterConfigFilePath(path string) string {
	return getPersisterConfigFilePath(path)
}

// GetDefaultDBConfig -
func GetDefaultDBConfig() *config.DBConfig {
	return &config.DBConfig{
		Type:              defaultType,
		BatchDelaySeconds: defaultBatchDelaySeconds,
		MaxBatchSize:      defaultMaxBatchSize,
		MaxOpenFiles:      defaultMaxOpenFiles,
	}
}
