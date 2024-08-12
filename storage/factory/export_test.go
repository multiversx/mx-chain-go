package factory

import (
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage"
)

// DefaultType exports the defaultType const to be used in tests
const DefaultType = defaultType

// DBConfigFileName exports the dbConfigFileName const to be used in tests
const DBConfigFileName = dbConfigFileName

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

// GetTmpFilePath -
func GetTmpFilePath(path string) (string, error) {
	return getTmpFilePath(path)
}
