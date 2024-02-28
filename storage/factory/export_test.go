package factory

import (
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage"
)

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

// NewPersisterCreator -
func NewPersisterCreator(config config.DBConfig) *persisterCreator {
	return newPersisterCreator(config)
}

// CreateShardIDProvider -
func (pc *persisterCreator) CreateShardIDProvider() (storage.ShardIDProvider, error) {
	return pc.createShardIDProvider()
}

// GetTmpFilePath -
func GetTmpFilePath(path string, pathSeparator string) (string, error) {
	return getTmpFilePath(path, pathSeparator)
}
