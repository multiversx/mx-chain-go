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
	"github.com/ElrondNetwork/elrond-go/factory/block"
	"github.com/ElrondNetwork/elrond-go/process/headerCheck"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/latestData"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

// BootstrapComponentsFactoryArgs holds the arguments needed to create a botstrap components factory
type BootstrapComponentsFactoryArgs struct {
	Config            config.Config
	EpochConfig       config.EpochConfig
	PrefConfig        config.Preferences
	ImportDbConfig    config.ImportDbConfig
	WorkingDir        string
	CoreComponents    CoreComponentsHolder
	CryptoComponents  CryptoComponentsHolder
	NetworkComponents NetworkComponentsHolder
}

type bootstrapComponentsFactory struct {
	config            config.Config
	epochConfig       config.EpochConfig
	prefConfig        config.Preferences
	importDbConfig    config.ImportDbConfig
	workingDir        string
	coreComponents    CoreComponentsHolder
	cryptoComponents  CryptoComponentsHolder
	networkComponents NetworkComponentsHolder
}

type bootstrapComponents struct {
	epochStartBootstrapper  EpochStartBootstrapper
	bootstrapParamsHolder   BootstrapParamsHolder
	nodeType                core.NodeType
	shardCoordinator        sharding.Coordinator
	headerVersionHandler    factory.HeaderVersionHandler
	versionedHeaderFactory  factory.VersionedHeaderFactory
	headerIntegrityVerifier factory.HeaderIntegrityVerifierHandler
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
		epochConfig:       args.EpochConfig,
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

	headerVersionHandler, err := block.NewHeaderVersionHandler(
		bcf.config.Versions.VersionsByEpochs,
		bcf.config.Versions.DefaultVersion,
		versionsCache,
	)
	if err != nil {
		return nil, err
	}

	headerIntegrityVerifier, err := headerCheck.NewHeaderIntegrityVerifier(
		[]byte(bcf.coreComponents.ChainID()),
		headerVersionHandler,
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

	latestStorageDataProvider, err := createLatestStorageDataProvider(
		bootstrapDataProvider,
		bcf.config,
		parentDir,
		core.DefaultEpochString,
		core.DefaultShardString,
	)
	if err != nil {
		return nil, err
	}

	unitOpener, err := createUnitOpener(
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
		EpochConfig:                bcf.epochConfig,
		EconomicsData:              bcf.coreComponents.EconomicsData(),
		GenesisNodesConfig:         bcf.coreComponents.GenesisNodesSetup(),
		GenesisShardCoordinator:    genesisShardCoordinator,
		StorageUnitOpener:          unitOpener,
		Rater:                      bcf.coreComponents.Rater(),
		DestinationShardAsObserver: destShardIdAsObserver,
		NodeShuffler:               bcf.coreComponents.NodesShuffler(),
		RoundHandler:               bcf.coreComponents.RoundHandler(),
		LatestStorageDataProvider:  latestStorageDataProvider,
		ArgumentsParser:            smartContract.NewArgumentParser(),
		StatusHandler:              bcf.coreComponents.StatusHandler(),
		HeaderIntegrityVerifier:    headerIntegrityVerifier,
	}

	var epochStartBootstrapper EpochStartBootstrapper
	if bcf.importDbConfig.IsImportDBMode {
		storageArg := bootstrap.ArgsStorageEpochStartBootstrap{
			ArgsEpochStartBootstrap:    epochStartBootstrapArgs,
			ImportDbConfig:             bcf.importDbConfig,
			ChanGracefullyClose:        bcf.coreComponents.ChanStopNodeProcess(),
			TimeToWaitForRequestedData: bootstrap.DefaultTimeToWaitForRequestedData,
		}

		epochStartBootstrapper, err = bootstrap.NewStorageEpochStartBootstrap(storageArg)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrNewStorageEpochStartBootstrap, err)
		}
	} else {
		epochStartBootstrapper, err = bootstrap.NewEpochStartBootstrap(epochStartBootstrapArgs)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrNewEpochStartBootstrap, err)
		}
	}

	bootstrapParameters, err := epochStartBootstrapper.Bootstrap()
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

	versionedHeaderFactory, err := bcf.createHeaderFactory(headerVersionHandler, bootstrapParameters.SelfShardId)
	if err != nil {
		return nil, err
	}

	return &bootstrapComponents{
		epochStartBootstrapper: epochStartBootstrapper,
		bootstrapParamsHolder: &bootstrapParams{
			bootstrapParams: bootstrapParameters,
		},
		nodeType:                nodeType,
		shardCoordinator:        shardCoordinator,
		headerVersionHandler:    headerVersionHandler,
		headerIntegrityVerifier: headerIntegrityVerifier,
		versionedHeaderFactory:  versionedHeaderFactory,
	}, nil
}

func (bcf *bootstrapComponentsFactory) createHeaderFactory(handler factory.HeaderVersionHandler, shardID uint32) (factory.VersionedHeaderFactory, error) {
	if shardID == core.MetachainShardId {
		return block.NewMetaHeaderFactory(handler)
	}
	return block.NewShardHeaderFactory(handler)
}

// Close closes the bootstrap components, closing at the same time any running goroutines
func (bc *bootstrapComponents) Close() error {
	// TODO: close all components
	if !check.IfNil(bc.epochStartBootstrapper) {
		return bc.epochStartBootstrapper.Close()
	}

	return nil
}

// NodeType returns the node type
func (bc *bootstrapComponents) NodeType() core.NodeType {
	return bc.nodeType
}

// ShardCoordinator returns the shard coordinator
func (bc *bootstrapComponents) ShardCoordinator() sharding.Coordinator {
	return bc.shardCoordinator
}

// HeaderVersionHandler returns the header version handler
func (bc *bootstrapComponents) HeaderVersionHandler() factory.HeaderVersionHandler {
	return bc.headerVersionHandler
}

// VersionedHeaderFactory returns the versioned header factory
func (bc *bootstrapComponents) VersionedHeaderFactory() factory.VersionedHeaderFactory {
	return bc.versionedHeaderFactory
}

// HeaderIntegrityVerifier returns the header integrity verifier
func (bc *bootstrapComponents) HeaderIntegrityVerifier() factory.HeaderIntegrityVerifierHandler {
	return bc.headerIntegrityVerifier
}

// createLatestStorageDataProvider will create a latest storage data provider handler
func createLatestStorageDataProvider(
	bootstrapDataProvider storageFactory.BootstrapDataProviderHandler,
	generalConfig config.Config,
	parentDir string,
	defaultEpochString string,
	defaultShardString string,
) (storage.LatestStorageDataProviderHandler, error) {
	directoryReader := storageFactory.NewDirectoryReader()

	latestStorageDataArgs := latestData.ArgsLatestDataProvider{
		GeneralConfig:         generalConfig,
		BootstrapDataProvider: bootstrapDataProvider,
		DirectoryReader:       directoryReader,
		ParentDir:             parentDir,
		DefaultEpochString:    defaultEpochString,
		DefaultShardString:    defaultShardString,
	}
	return latestData.NewLatestDataProvider(latestStorageDataArgs)
}

// createUnitOpener will create a new unit opener handler
func createUnitOpener(
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
