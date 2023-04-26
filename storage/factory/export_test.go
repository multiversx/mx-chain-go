package factory

import "github.com/multiversx/mx-chain-go/config"

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
