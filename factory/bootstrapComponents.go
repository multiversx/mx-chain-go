package factory

import (
	"fmt"
	"path/filepath"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/process/headerCheck"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

// BootstrapComponentsFactoryArgs holds the arguments needed to create a botstrap components factory
type BootstrapComponentsFactoryArgs struct {
	Config            config.Config
	PrefConfig        config.Preferences
	ImportDbConfig    config.ImportDbConfig
	WorkingDir        string
	CoreComponents    CoreComponentsHolder
	CryptoComponents  CryptoComponentsHolder
	NetworkComponents NetworkComponentsHolder
}

type bootstrapComponentsFactory struct {
	config            config.Config
	prefConfig        config.Preferences
	importDbConfig    config.ImportDbConfig
	workingDir        string
	coreComponents    CoreComponentsHolder
	cryptoComponents  CryptoComponentsHolder
	networkComponents NetworkComponentsHolder
}

type bootstrapParamsHolder struct {
	bootstrapParams bootstrap.Parameters
}

type bootstrapComponents struct {
	epochStartBootstraper   EpochStartBootstrapper
	bootstrapParamsHandler  BootstrapParamsHandler
	nodeType                core.NodeType
	shardCoordinator        sharding.Coordinator
	headerIntegrityVerifier HeaderIntegrityVerifierHandler
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
	if args.WorkingDir == "" {
		return nil, errors.ErrInvalidWorkingDir
	}

	return &bootstrapComponentsFactory{
		config:            args.Config,
		prefConfig:        args.PrefConfig,
		importDbConfig:    args.ImportDbConfig,
		workingDir:        args.WorkingDir,
		coreComponents:    args.CoreComponents,
		cryptoComponents:  args.CryptoComponents,
		networkComponents: args.NetworkComponents,
	}, nil
}

// Create creates the bootstrap components
func (bcf *bootstrapComponentsFactory) Create() (*bootstrapComponents, error) {
	destShardIdAsObserver, err := core.ProcessDestinationShardAsObserver(bcf.prefConfig.Preferences.DestinationShardAsObserver)
	if err != nil {
		return nil, err
	}

	versionsCache, err := storageUnit.NewCache(storageFactory.GetCacherFromConfig(bcf.config.Versions.Cache))
	if err != nil {
		return nil, err
	}

	headerIntegrityVerifier, err := headerCheck.NewHeaderIntegrityVerifier(
		[]byte(bcf.coreComponents.ChainID()),
		bcf.config.Versions.VersionsByEpochs,
		bcf.config.Versions.DefaultVersion,
		versionsCache,
	)
	if err != nil {
		return nil, err
	}
	genesisShardCoordinator, nodeType, err := CreateShardCoordinator(
		bcf.coreComponents.GenesisNodesSetup(),
		bcf.cryptoComponents.PublicKey(),
		bcf.prefConfig.Preferences,
		log,
	)
	if err != nil {
		return nil, err
	}

	bootstrapDataProvider, err := storageFactory.NewBootstrapDataProvider(bcf.coreComponents.InternalMarshalizer())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrNewBootstrapDataProvider, err)
	}

	parentDir := filepath.Join(
		bcf.workingDir,
		core.DefaultDBPath,
		bcf.coreComponents.ChainID())

	latestStorageDataProvider, err := CreateLatestStorageDataProvider(
		bootstrapDataProvider,
		bcf.config,
		parentDir,
		core.DefaultEpochString,
		core.DefaultShardString,
	)
	if err != nil {
		return nil, err
	}

	unitOpener, err := CreateUnitOpener(
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
		GenesisShardCoordinator:    genesisShardCoordinator,
		StorageUnitOpener:          unitOpener,
		Rater:                      bcf.coreComponents.Rater(),
		DestinationShardAsObserver: destShardIdAsObserver,
		NodeShuffler:               bcf.coreComponents.NodesShuffler(),
		Rounder:                    bcf.coreComponents.Rounder(),
		LatestStorageDataProvider:  latestStorageDataProvider,
		ArgumentsParser:            smartContract.NewArgumentParser(),
		StatusHandler:              bcf.coreComponents.StatusHandler(),
		HeaderIntegrityVerifier:    headerIntegrityVerifier,
	}

	var epochStartBootstraper EpochStartBootstrapper
	if bcf.importDbConfig.IsImportDBMode {
		storageArg := bootstrap.ArgsStorageEpochStartBootstrap{
			ArgsEpochStartBootstrap:    epochStartBootstrapArgs,
			ImportDbConfig:             bcf.importDbConfig,
			ChanGracefullyClose:        bcf.coreComponents.ChanStopNodeProcess(),
			TimeToWaitForRequestedData: bootstrap.DefaultTimeToWaitForRequestedData,
		}

		epochStartBootstraper, err = bootstrap.NewStorageEpochStartBootstrap(storageArg)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrNewStorageEpochStartBootstrap, err)
		}
	} else {
		epochStartBootstraper, err = bootstrap.NewEpochStartBootstrap(epochStartBootstrapArgs)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrNewEpochStartBootstrap, err)
		}
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

	shardCoordinator, err := sharding.NewMultiShardCoordinator(
		bootstrapParameters.NumOfShards,
		bootstrapParameters.SelfShardId)
	if err != nil {
		return nil, err
	}

	return &bootstrapComponents{
		epochStartBootstraper: epochStartBootstraper,
		bootstrapParamsHandler: &bootstrapParamsHolder{
			bootstrapParams: bootstrapParameters,
		},
		nodeType:                nodeType,
		shardCoordinator:        shardCoordinator,
		headerIntegrityVerifier: headerIntegrityVerifier,
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

// NodeType returns the node type
func (bph *bootstrapComponents) NodeType() core.NodeType {
	return bph.nodeType
}

// ShardCoordinator returns the shard coordinator
func (bph *bootstrapComponents) ShardCoordinator() sharding.Coordinator {
	return bph.shardCoordinator
}

// HeaderIntegrityVerifier returns the header integrity verifier
func (bph *bootstrapComponents) HeaderIntegrityVerifier() HeaderIntegrityVerifierHandler {
	return bph.headerIntegrityVerifier
}

// IsInterfaceNil returns true if the underlying object is nil
func (bph *bootstrapParamsHolder) IsInterfaceNil() bool {
	return bph == nil
}

// CreateLatestStorageDataProvider will create a latest storage data provider handler
func CreateLatestStorageDataProvider(
	bootstrapDataProvider storageFactory.BootstrapDataProviderHandler,
	generalConfig config.Config,
	parentDir string,
	defaultEpochString string,
	defaultShardString string,
) (storage.LatestStorageDataProviderHandler, error) {
	directoryReader := storageFactory.NewDirectoryReader()

	latestStorageDataArgs := storageFactory.ArgsLatestDataProvider{
		GeneralConfig:         generalConfig,
		BootstrapDataProvider: bootstrapDataProvider,
		DirectoryReader:       directoryReader,
		ParentDir:             parentDir,
		DefaultEpochString:    defaultEpochString,
		DefaultShardString:    defaultShardString,
	}
	return storageFactory.NewLatestDataProvider(latestStorageDataArgs)
}

// CreateUnitOpener will create a new unit opener handler
func CreateUnitOpener(
	bootstrapDataProvider storageFactory.BootstrapDataProviderHandler,
	latestDataFromStorageProvider storage.LatestStorageDataProviderHandler,
	generalConfig config.Config,
	defaultEpochString string,
	defaultShardString string,
) (storage.UnitOpenerHandler, error) {
	argsStorageUnitOpener := storageFactory.ArgsNewOpenStorageUnits{
		GeneralConfig:             generalConfig,
		BootstrapDataProvider:     bootstrapDataProvider,
		LatestStorageDataProvider: latestDataFromStorageProvider,
		DefaultEpochString:        defaultEpochString,
		DefaultShardString:        defaultShardString,
	}

	return storageFactory.NewStorageUnitOpenHandler(argsStorageUnitOpener)
}
