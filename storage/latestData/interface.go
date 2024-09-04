package latestData

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/storage"
)

// LatestDataProviderFactory defines the factory for latest data provider
type LatestDataProviderFactory interface {
	CreateLatestDataProvider(args ArgsLatestDataProvider) (storage.LatestStorageDataProviderHandler, error)
	IsInterfaceNil() bool
}

// MetaEpochStartTriggerRegistryHandler defines the interface needed to create registry trigger for epoch start
type MetaEpochStartTriggerRegistryHandler interface {
	UnmarshalTrigger(marshaller marshal.Marshalizer, data []byte) (data.MetaTriggerRegistryHandler, error)
}

type epochStartRoundLoaderHandler interface {
	loadEpochStartRound(
		shardID uint32,
		key []byte,
		storer storage.Storer,
	) (uint64, error)
}
