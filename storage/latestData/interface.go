package latestData

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/storage"
)

type LatestDataProviderFactory interface {
	CreateLatestDataProvider(args ArgsLatestDataProvider) (storage.LatestStorageDataProviderHandler, error)
	IsInterfaceNil() bool
}

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
