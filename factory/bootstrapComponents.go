package factory

import (
	"fmt"
	"path/filepath"

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
	if check.IfNil(args.ShardCoordinator) {
		return nil, errors.ErrNilShardCoordinator
	}
	if check.IfNil(args.HeaderIntegrityVerifier) {
		return nil, errors.ErrNilHeaderIntegrityVerifier
	}
	if args.WorkingDir == "" {
		return nil, errors.ErrInvalidWorkingDir
	}

	return &bootstrapComponentsFactory{
		config:                  args.Config,
		workingDir:              args.WorkingDir,
		destinationAsObserver:   args.DestinationAsObserver,
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
		return nil, fmt.Errorf("%w: %v", errors.ErrNewBootstrapDataProvider, err)
	}

	parentDir := filepath.Join(
		bcf.workingDir,
		core.DefaultDBPath,
		bcf.coreComponents.ChainID())

	latestStorageDataProvider, err := factory.CreateLatestStorageDataProvider(
		bootstrapDataProvider,
		bcf.config,
		parentDir,
		core.DefaultEpochString,
		core.DefaultShardString,
	)
	if err != nil {
		return nil, err
	}

	unitOpener, err := factory.CreateUnitOpener(
		bootstrapDataProvider,
		latestStorageDataProvider,
		bcf.config,
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
		GenesisNodesConfig:         bcf.coreComponents.GenesisNodesSetup(),
		GenesisShardCoordinator:    bcf.shardCoordinator,
		StorageUnitOpener:          unitOpener,
		Rater:                      bcf.coreComponents.Rater(),
		DestinationShardAsObserver: bcf.destinationAsObserver,
		NodeShuffler:               bcf.coreComponents.NodesShuffler(),
		Rounder:                    bcf.coreComponents.Rounder(),
		LatestStorageDataProvider:  latestStorageDataProvider,
		ArgumentsParser:            smartContract.NewArgumentParser(),
		StatusHandler:              bcf.coreComponents.StatusHandler(),
		HeaderIntegrityVerifier:    bcf.headerIntegrityVerifier,
	}

	epochStartBootstraper, err := bootstrap.NewEpochStartBootstrap(epochStartBootstrapArgs)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrNewEpochStartBootstrap, err)
	}

	bootstrapParameters, err := epochStartBootstraper.Bootstrap()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrBootstrap, err)
	}

	log.Info("bootstrap parameters",
		"shardId", bootstrapParameters.SelfShardId,
		"epoch", bootstrapParameters.Epoch,
		"numShards", bootstrapParameters.NumOfShards,
	)

	return &bootstrapComponents{
		epochStartBootstraper: epochStartBootstraper,
		bootstrapParamsHandler: &bootstrapParamsHolder{
			bootstrapParams: bootstrapParameters,
		},
	}, nil
}

// Close closes the bootstrap components, closing at the same time any running goroutines
func (bc *bootstrapComponents) Close() error {
	// TODO: close all components
	if !check.IfNil(bc.epochStartBootstraper) {
		return bc.epochStartBootstraper.Close()
	}

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
