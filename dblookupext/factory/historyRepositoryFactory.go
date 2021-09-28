package factory

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dblookupext"
	"github.com/ElrondNetwork/elrond-go/dblookupext/disabled"
	"github.com/ElrondNetwork/elrond-go/dblookupext/esdtSupply"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ArgsHistoryRepositoryFactory holds all dependencies required by the history processor factory in order to create
// new instances
type ArgsHistoryRepositoryFactory struct {
	SelfShardID              uint32
	Config                   config.DbLookupExtensionsConfig
	Store                    dataRetriever.StorageService
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	Uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
}

type historyRepositoryFactory struct {
	selfShardID              uint32
	dbLookupExtensionsConfig config.DbLookupExtensionsConfig
	store                    dataRetriever.StorageService
	marshalizer              marshal.Marshalizer
	hasher                   hashing.Hasher
	uInt64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
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
	if check.IfNil(args.Uint64ByteSliceConverter) {
		return nil, process.ErrNilUint64Converter
	}

	return &historyRepositoryFactory{
		selfShardID:              args.SelfShardID,
		dbLookupExtensionsConfig: args.Config,
		store:                    args.Store,
		marshalizer:              args.Marshalizer,
		hasher:                   args.Hasher,
		uInt64ByteSliceConverter: args.Uint64ByteSliceConverter,
	}, nil
}

// Create creates instances of HistoryRepository
func (hpf *historyRepositoryFactory) Create() (dblookupext.HistoryRepository, error) {
	if !hpf.dbLookupExtensionsConfig.Enabled {
		return disabled.NewNilHistoryRepository()
	}

	esdtSuppliesHandler, err := esdtSupply.NewSuppliesProcessor(
		hpf.marshalizer,
		hpf.store.GetStorer(dataRetriever.ESDTSuppliesUnit),
		hpf.store.GetStorer(dataRetriever.TxLogsUnit),
	)
	if err != nil {
		return nil, err
	}
	blockHashByRoundDataUnit := dataRetriever.ShardHdrRoundHashDataUnit + dataRetriever.UnitType(hpf.selfShardID)
	if hpf.selfShardID == core.MetachainShardId {
		blockHashByRoundDataUnit = dataRetriever.MetaHdrRoundHashDataUnit
	}

	historyRepArgs := dblookupext.HistoryRepositoryArguments{
		SelfShardID:                 hpf.selfShardID,
		Hasher:                      hpf.hasher,
		Marshalizer:                 hpf.marshalizer,
		BlockHashByRound:            hpf.store.GetStorer(blockHashByRoundDataUnit),
		Uint64ByteSliceConverter:    hpf.uInt64ByteSliceConverter,
		MiniblocksMetadataStorer:    hpf.store.GetStorer(dataRetriever.MiniblocksMetadataUnit),
		EpochByHashStorer:           hpf.store.GetStorer(dataRetriever.EpochByHashUnit),
		MiniblockHashByTxHashStorer: hpf.store.GetStorer(dataRetriever.MiniblockHashByTxHashUnit),
		EventsHashesByTxHashStorer:  hpf.store.GetStorer(dataRetriever.ResultsHashesByTxHashUnit),
		ESDTSuppliesHandler:         esdtSuppliesHandler,
	}
	return dblookupext.NewHistoryRepository(historyRepArgs)
}

// IsInterfaceNil returns true if there is no value under the interface
func (hpf *historyRepositoryFactory) IsInterfaceNil() bool {
	return hpf == nil
}
