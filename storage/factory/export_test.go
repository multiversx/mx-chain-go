package factory

import (
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage"
)

// GetPersisterConfigFilePath -
func GetPersisterConfigFilePath(path string) string {
	return getPersisterConfigFilePath(path)
}

// NewPersisterCreator -
func NewPersisterCreator(config config.DBConfig) *persisterCreator {
	return newPersisterCreator(config)
}

// CreateShardIDProvider -
func (pc *persisterCreator) CreateShardIDProvider() (storage.ShardIDProvider, error) {
	return pc.createShardIDProvider()
}
