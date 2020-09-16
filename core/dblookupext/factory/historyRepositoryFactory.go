package factory

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/dblookupext"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

// ArgsHistoryRepositoryFactory holds all dependencies required by the history processor factory in order to create
// new instances
type ArgsHistoryRepositoryFactory struct {
	SelfShardID uint32
	Config      config.DbLookupExtensionsConfig
	Store       dataRetriever.StorageService
	Marshalizer marshal.Marshalizer
	Hasher      hashing.Hasher
}

type historyRepositoryFactory struct {
	selfShardID              uint32
	dbLookupExtensionsConfig config.DbLookupExtensionsConfig
	store                    dataRetriever.StorageService
	marshalizer              marshal.Marshalizer
	hasher                   hashing.Hasher
}

// NewHistoryRepositoryFactory creates an instance of historyRepositoryFactory
func NewHistoryRepositoryFactory(args *ArgsHistoryRepositoryFactory) (dblookupext.HistoryRepositoryFactory, error) {
	if check.IfNil(args.Marshalizer) {
		return nil, core.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, core.ErrNilHasher
	}
	if check.IfNil(args.Store) {
		return nil, core.ErrNilStore
	}

	return &historyRepositoryFactory{
		selfShardID:              args.SelfShardID,
		dbLookupExtensionsConfig: args.Config,
		store:                    args.Store,
		marshalizer:              args.Marshalizer,
		hasher:                   args.Hasher,
	}, nil
}

// Create creates instances of HistoryRepository
func (hpf *historyRepositoryFactory) Create() (dblookupext.HistoryRepository, error) {
	if !hpf.dbLookupExtensionsConfig.Enabled {
		return dblookupext.NewNilHistoryRepository()
	}

	historyRepArgs := dblookupext.HistoryRepositoryArguments{
		SelfShardID:                 hpf.selfShardID,
		Hasher:                      hpf.hasher,
		Marshalizer:                 hpf.marshalizer,
		MiniblocksMetadataStorer:    hpf.store.GetStorer(dataRetriever.MiniblocksMetadataUnit),
		EpochByHashStorer:           hpf.store.GetStorer(dataRetriever.EpochByHashUnit),
		MiniblockHashByTxHashStorer: hpf.store.GetStorer(dataRetriever.MiniblockHashByTxHashUnit),
	}
	return dblookupext.NewHistoryRepository(historyRepArgs)
}

// IsInterfaceNil returns true if there is no value under the interface
func (hpf *historyRepositoryFactory) IsInterfaceNil() bool {
	return hpf == nil
}
