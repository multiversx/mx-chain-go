package factory

import (
	"github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/sharding"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
)

// BootstrapComponentsFactoryArgs holds the arguments needed to create a botstrap components factory
type BootstrapComponentsFactoryArgs struct {
	Config                  config.Config
	WorkingDir              string
	DestinationAsObserver   uint32
	GenesisNodesSetup       sharding.GenesisNodesSetupHandler
	NodeShuffler            sharding.NodesShuffler
	ShardCoordinator        sharding.Coordinator
	CoreComponents          CoreComponentsHolder
	CryptoComponents        CryptoComponentsHolder
	NetworkComponents       NetworkComponentsHolder
	HeaderIntegrityVerifier process.HeaderIntegrityVerifier
}

type bootstrapComponentsFactory struct {
	config                  config.Config
	workingDir              string
	destinationAsObserver   uint32
	genesisNodesSetup       sharding.GenesisNodesSetupHandler
	nodesShuffler           sharding.NodesShuffler
	shardCoordinator        sharding.Coordinator
	coreComponents          CoreComponentsHolder
	cryptoComponents        CryptoComponentsHolder
	networkComponents       NetworkComponentsHolder
	headerIntegrityVerifier process.HeaderIntegrityVerifier
}

type bootstrapParamsHolder struct {
	bootstrapParams bootstrap.Parameters
}

type bootstrapComponents struct {
	epochStartBootstraper  EpochStartBootstrapper
	bootstrapParamsHandler BootstrapParamsHandler
}

// NewBootstrapComponentsFactory creates an instance of bootstrapComponentsFactory
func NewBootstrapComponentsFactory(args BootstrapComponentsFactoryArgs) (*bootstrapComponentsFactory, error) {
	if check.IfNil(args.CoreComponents) {
		return nil, errors.ErrNilCoreComponentsHolder
	}
	if check.IfNil(args.CryptoComponents) {
		return nil, errors.ErrNilCryptoComponentsHolder
	}
	if check.IfNil(args.NetworkComponents) {
		return nil, errors.ErrNilNetworkComponentsHolder
	}
	if check.IfNil(args.NodeShuffler) {
		return nil, errors.ErrNilShuffler
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, errors.ErrNilShardCoordinator
	}
	if check.IfNil(args.GenesisNodesSetup) {
		return nil, errors.ErrNilGenesisNodesSetup
	}
	if args.WorkingDir == "" {
		return nil, errors.ErrInvalidWorkingDir
	}

	return &bootstrapComponentsFactory{
		config:                  args.Config,
		workingDir:              args.WorkingDir,
		destinationAsObserver:   args.DestinationAsObserver,
		genesisNodesSetup:       args.GenesisNodesSetup,
		nodesShuffler:           args.NodeShuffler,
		shardCoordinator:        args.ShardCoordinator,
		coreComponents:          args.CoreComponents,
		cryptoComponents:        args.CryptoComponents,
		networkComponents:       args.NetworkComponents,
		headerIntegrityVerifier: args.HeaderIntegrityVerifier,
	}, nil
}

// Create creates the bootstrap components
func (bcf *bootstrapComponentsFactory) Create() (*bootstrapComponents, error) {
	var err error

	bootstrapDataProvider, err := storageFactory.NewBootstrapDataProvider(bcf.coreComponents.InternalMarshalizer())
	if err != nil {
		return nil, err
	}

	latestStorageDataProvider, err := factory.CreateLatestStorageDataProvider(
		bootstrapDataProvider,
		bcf.coreComponents.InternalMarshalizer(),
		bcf.coreComponents.Hasher(),
		bcf.config,
		bcf.coreComponents.ChainID(),
		bcf.workingDir,
		core.DefaultDBPath,
		core.DefaultEpochString,
		core.DefaultShardString,
	)
	if err != nil {
		return nil, err
	}

	unitOpener, err := factory.CreateUnitOpener(
		bootstrapDataProvider,
		latestStorageDataProvider,
		bcf.coreComponents.InternalMarshalizer(),
		bcf.config,
		bcf.coreComponents.ChainID(),
		bcf.workingDir,
		core.DefaultDBPath,
		core.DefaultEpochString,
		core.DefaultShardString,
	)
	if err != nil {
		return nil, err
	}

	epochStartBootstrapArgs := bootstrap.ArgsEpochStartBootstrap{
		CoreComponentsHolder:       bcf.coreComponents,
		CryptoComponentsHolder:     bcf.cryptoComponents,
		Messenger:                  bcf.networkComponents.NetworkMessenger(),
		GeneralConfig:              bcf.config,
		EconomicsData:              bcf.coreComponents.EconomicsData(),
		GenesisNodesConfig:         bcf.genesisNodesSetup,
		GenesisShardCoordinator:    bcf.shardCoordinator,
		StorageUnitOpener:          unitOpener,
		Rater:                      bcf.coreComponents.Rater(),
		DestinationShardAsObserver: bcf.destinationAsObserver,
		NodeShuffler:               bcf.nodesShuffler,
		Rounder:                    bcf.coreComponents.Rounder(),
		LatestStorageDataProvider:  latestStorageDataProvider,
		ArgumentsParser:            smartContract.NewArgumentParser(),
		StatusHandler:              bcf.coreComponents.StatusHandler(),
		HeaderIntegrityVerifier:    bcf.headerIntegrityVerifier,
	}

	epochStartBootstraper, err := bootstrap.NewEpochStartBootstrap(epochStartBootstrapArgs)
	if err != nil {
		log.Error("could not create bootstrap", "err", err)
		return nil, err
	}

	bootstrapParameters, err := epochStartBootstraper.Bootstrap()
	if err != nil {
		log.Error("bootstrap return error", "error", err)
		return nil, err
	}

	log.Info("bootstrap parameters", "shardId", bootstrapParameters.SelfShardId, "epoch", bootstrapParameters.Epoch, "numShards", bootstrapParameters.NumOfShards)

	return &bootstrapComponents{
		epochStartBootstraper: epochStartBootstraper,
		bootstrapParamsHandler: &bootstrapParamsHolder{
			bootstrapParams: bootstrapParameters,
		},
	}, nil
}

// Close closes the bootstrap components, closing at the same time any running goroutines
func (bc *bootstrapComponents) Close() error {
	return nil
}

// Epoch returns the epoch number after bootstrap
func (bph *bootstrapParamsHolder) Epoch() uint32 {
	return bph.bootstrapParams.Epoch
}

// SelfShardID returns the self shard ID after bootstrap
func (bph *bootstrapParamsHolder) SelfShardID() uint32 {
	return bph.bootstrapParams.SelfShardId
}

// NumOfShards returns the number of shards after bootstrap
func (bph *bootstrapParamsHolder) NumOfShards() uint32 {
	return bph.bootstrapParams.NumOfShards
}

// NodesConfig returns the nodes coordinator config after bootstrap
func (bph *bootstrapParamsHolder) NodesConfig() *sharding.NodesCoordinatorRegistry {
	return bph.bootstrapParams.NodesConfig
}

// IsInterfaceNil returns true if the underlying object is nil
func (bph *bootstrapParamsHolder) IsInterfaceNil() bool {
	return bph == nil
}
