package factory

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/blockchain"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	dataRetrieverFactory "github.com/ElrondNetwork/elrond-go/dataRetriever/factory"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/provider"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage/factory"
)

// DataComponentsFactoryArgs holds the arguments needed for creating a data components factory
type DataComponentsFactoryArgs struct {
	Config                        config.Config
	ShardCoordinator              sharding.Coordinator
	Core                          CoreComponentsHolder
	EpochStartNotifier            EpochStartNotifier
	CurrentEpoch                  uint32
	CreateTrieEpochRootHashStorer bool
}

type dataComponentsFactory struct {
	config                        config.Config
	shardCoordinator              sharding.Coordinator
	core                          CoreComponentsHolder
	epochStartNotifier            EpochStartNotifier
	currentEpoch                  uint32
	createTrieEpochRootHashStorer bool
}

// dataComponents struct holds the data components
type dataComponents struct {
	blkc               data.ChainHandler
	store              dataRetriever.StorageService
	datapool           dataRetriever.PoolsHolder
	miniBlocksProvider MiniBlockProvider
}

// NewDataComponentsFactory will return a new instance of dataComponentsFactory
func NewDataComponentsFactory(args DataComponentsFactoryArgs) (*dataComponentsFactory, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, errors.ErrNilShardCoordinator
	}
	if check.IfNil(args.Core) {
		return nil, errors.ErrNilCoreComponents
	}
	if check.IfNil(args.Core.PathHandler()) {
		return nil, errors.ErrNilPathHandler
	}
	if check.IfNil(args.EpochStartNotifier) {
		return nil, errors.ErrNilEpochStartNotifier
	}
	if check.IfNil(args.Core.EconomicsData()) {
		return nil, errors.ErrNilEconomicsHandler
	}

	return &dataComponentsFactory{
		config:                        args.Config,
		shardCoordinator:              args.ShardCoordinator,
		core:                          args.Core,
		epochStartNotifier:            args.EpochStartNotifier,
		currentEpoch:                  args.CurrentEpoch,
		createTrieEpochRootHashStorer: args.CreateTrieEpochRootHashStorer,
	}, nil
}

// Create will create and return the data components
func (dcf *dataComponentsFactory) Create() (*dataComponents, error) {
	var datapool dataRetriever.PoolsHolder
	blkc, err := dcf.createBlockChainFromConfig()
	if err != nil {
		return nil, err
	}

	store, err := dcf.createDataStoreFromConfig()
	if err != nil {
		return nil, err
	}

	dataPoolArgs := dataRetrieverFactory.ArgsDataPool{
		Config:           &dcf.config,
		EconomicsData:    dcf.core.EconomicsData(),
		ShardCoordinator: dcf.shardCoordinator,
	}
	datapool, err = dataRetrieverFactory.NewDataPoolFromConfig(dataPoolArgs)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errors.ErrDataPoolCreation, err.Error())
	}

	arg := provider.ArgMiniBlockProvider{
		MiniBlockPool:    datapool.MiniBlocks(),
		MiniBlockStorage: store.GetStorer(dataRetriever.MiniBlockUnit),
		Marshalizer:      dcf.core.InternalMarshalizer(),
	}

	miniBlocksProvider, err := provider.NewMiniBlockProvider(arg)
	if err != nil {
		return nil, err
	}

	return &dataComponents{
		blkc:               blkc,
		store:              store,
		datapool:           datapool,
		miniBlocksProvider: miniBlocksProvider,
	}, nil
}

func (dcf *dataComponentsFactory) createBlockChainFromConfig() (data.ChainHandler, error) {
	if dcf.shardCoordinator.SelfId() < dcf.shardCoordinator.NumberOfShards() {
		blockChain, err := blockchain.NewBlockChain(dcf.core.StatusHandler())
		if err != nil {
			return nil, err
		}
		return blockChain, nil
	}
	if dcf.shardCoordinator.SelfId() == core.MetachainShardId {
		blockChain, err := blockchain.NewMetaChain(dcf.core.StatusHandler())
		if err != nil {
			return nil, err
		}
		return blockChain, nil
	}
	return nil, errors.ErrBlockchainCreation
}

func (dcf *dataComponentsFactory) createDataStoreFromConfig() (dataRetriever.StorageService, error) {
	storageServiceFactory, err := factory.NewStorageServiceFactory(
		&dcf.config,
		dcf.shardCoordinator,
		dcf.core.PathHandler(),
		dcf.epochStartNotifier,
		dcf.currentEpoch,
		dcf.createTrieEpochRootHashStorer,
	)
	if err != nil {
		return nil, err
	}
	if dcf.shardCoordinator.SelfId() < dcf.shardCoordinator.NumberOfShards() {
		return storageServiceFactory.CreateForShard()
	}
	if dcf.shardCoordinator.SelfId() == core.MetachainShardId {
		return storageServiceFactory.CreateForMeta()
	}
	return nil, errors.ErrDataStoreCreation
}

// Close closes all underlying components that need closing
func (cc *dataComponents) Close() error {
	if cc.store != nil {
		log.Debug("closing all store units....")
		return cc.store.CloseAll()
	}

	return nil
}
