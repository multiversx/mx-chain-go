package data

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/blockchain"
	dataRetrieverFactory "github.com/multiversx/mx-chain-go/dataRetriever/factory"
	"github.com/multiversx/mx-chain-go/dataRetriever/provider"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/sharding"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// DataComponentsFactoryArgs holds the arguments needed for creating a data components factory
type DataComponentsFactoryArgs struct {
	Config                        config.Config
	PrefsConfig                   config.PreferencesConfig
	ShardCoordinator              sharding.Coordinator
	Core                          factory.CoreComponentsHolder
	StatusCore                    factory.StatusCoreComponentsHolder
	Crypto                        factory.CryptoComponentsHolder
	FlagsConfigs                  config.ContextFlagsConfig
	CurrentEpoch                  uint32
	CreateTrieEpochRootHashStorer bool
	NodeProcessingMode            common.NodeProcessingMode
}

type dataComponentsFactory struct {
	config                        config.Config
	prefsConfig                   config.PreferencesConfig
	shardCoordinator              sharding.Coordinator
	core                          factory.CoreComponentsHolder
	statusCore                    factory.StatusCoreComponentsHolder
	crypto                        factory.CryptoComponentsHolder
	flagsConfig                   config.ContextFlagsConfig
	currentEpoch                  uint32
	createTrieEpochRootHashStorer bool
	nodeProcessingMode            common.NodeProcessingMode
}

// dataComponents struct holds the data components
type dataComponents struct {
	blkc               data.ChainHandler
	store              dataRetriever.StorageService
	datapool           dataRetriever.PoolsHolder
	miniBlocksProvider factory.MiniBlockProvider
}

var log = logger.GetOrCreate("factory")

// NewDataComponentsFactory will return a new instance of dataComponentsFactory
func NewDataComponentsFactory(args DataComponentsFactoryArgs) (*dataComponentsFactory, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, errors.ErrNilShardCoordinator
	}
	if check.IfNil(args.Core) {
		return nil, errors.ErrNilCoreComponents
	}
	if check.IfNil(args.StatusCore) {
		return nil, errors.ErrNilStatusCoreComponents
	}
	if check.IfNil(args.Crypto) {
		return nil, errors.ErrNilCryptoComponents
	}

	return &dataComponentsFactory{
		config:                        args.Config,
		prefsConfig:                   args.PrefsConfig,
		shardCoordinator:              args.ShardCoordinator,
		core:                          args.Core,
		statusCore:                    args.StatusCore,
		currentEpoch:                  args.CurrentEpoch,
		createTrieEpochRootHashStorer: args.CreateTrieEpochRootHashStorer,
		flagsConfig:                   args.FlagsConfigs,
		nodeProcessingMode:            args.NodeProcessingMode,
		crypto:                        args.Crypto,
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
		Marshalizer:      dcf.core.InternalMarshalizer(),
		PathManager:      dcf.core.PathHandler(),
		PersisterFactory: dcf.core.PersisterFactory(),
	}
	datapool, err = dataRetrieverFactory.NewDataPoolFromConfig(dataPoolArgs)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errors.ErrDataPoolCreation, err.Error())
	}

	log.Debug("closing the datapool trie nodes cacher")
	errNotCritical := datapool.TrieNodes().Close()
	if errNotCritical != nil {
		log.Warn("unable to close the trie nodes cacher...continuing", "error", errNotCritical)
	}

	miniBlockStorer, err := store.GetStorer(dataRetriever.MiniBlockUnit)
	if err != nil {
		return nil, err
	}

	arg := provider.ArgMiniBlockProvider{
		MiniBlockPool:    datapool.MiniBlocks(),
		MiniBlockStorage: miniBlockStorer,
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
		blockChain, err := blockchain.NewBlockChain(dcf.statusCore.AppStatusHandler())
		if err != nil {
			return nil, err
		}
		return blockChain, nil
	}
	if dcf.shardCoordinator.SelfId() == core.MetachainShardId {
		blockChain, err := blockchain.NewMetaChain(dcf.statusCore.AppStatusHandler())
		if err != nil {
			return nil, err
		}
		return blockChain, nil
	}
	return nil, errors.ErrBlockchainCreation
}

func (dcf *dataComponentsFactory) createDataStoreFromConfig() (dataRetriever.StorageService, error) {
	storageServiceFactory, err := storageFactory.NewStorageServiceFactory(
		storageFactory.StorageServiceFactoryArgs{
			Config:                        dcf.config,
			PrefsConfig:                   dcf.prefsConfig,
			ShardCoordinator:              dcf.shardCoordinator,
			PathManager:                   dcf.core.PathHandler(),
			EpochStartNotifier:            dcf.core.EpochStartNotifierWithConfirm(),
			NodeTypeProvider:              dcf.core.NodeTypeProvider(),
			CurrentEpoch:                  dcf.currentEpoch,
			StorageType:                   storageFactory.ProcessStorageService,
			CreateTrieEpochRootHashStorer: dcf.createTrieEpochRootHashStorer,
			NodeProcessingMode:            dcf.nodeProcessingMode,
			RepopulateTokensSupplies:      dcf.flagsConfig.RepopulateTokensSupplies,
			ManagedPeersHolder:            dcf.crypto.ManagedPeersHolder(),
			StateStatsHandler:             dcf.statusCore.StateStatsHandler(),
			PersisterFactory:              dcf.core.PersisterFactory(),
		})
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
	var lastError error
	if cc.store != nil {
		log.Debug("closing all store units....")
		err := cc.store.CloseAll()
		if err != nil {
			log.Error("failed to close all store units", "error", err.Error())
			lastError = err
		}
	}

	if !check.IfNil(cc.datapool) {
		lastError = cc.datapool.Close()
	}

	return lastError
}
