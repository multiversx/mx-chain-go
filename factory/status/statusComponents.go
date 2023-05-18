package status

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	nodeData "github.com/multiversx/mx-chain-core-go/data"
	outportCore "github.com/multiversx/mx-chain-core-go/data/outport"
	factoryMarshalizer "github.com/multiversx/mx-chain-core-go/marshal/factory"
	indexerFactory "github.com/multiversx/mx-chain-es-indexer-go/process/factory"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/statistics"
	swVersionFactory "github.com/multiversx/mx-chain-go/common/statistics/softwareVersion/factory"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/notifier"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/outport"
	outportDriverFactory "github.com/multiversx/mx-chain-go/outport/factory"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type statusComponents struct {
	nodesCoordinator nodesCoordinator.NodesCoordinator
	statusHandler    core.AppStatusHandler
	outportHandler   outport.OutportHandler
	softwareVersion  statistics.SoftwareVersionChecker
	cancelFunc       func()
}

// StatusComponentsFactoryArgs redefines the arguments structure needed for the status components factory
type StatusComponentsFactoryArgs struct {
	Config               config.Config
	ExternalConfig       config.ExternalConfig
	EconomicsConfig      config.EconomicsConfig
	ShardCoordinator     sharding.Coordinator
	NodesCoordinator     nodesCoordinator.NodesCoordinator
	EpochStartNotifier   factory.EpochStartNotifier
	CoreComponents       factory.CoreComponentsHolder
	StatusCoreComponents factory.StatusCoreComponentsHolder
	NetworkComponents    factory.NetworkComponentsHolder
	StateComponents      factory.StateComponentsHolder
	IsInImportMode       bool
}

type statusComponentsFactory struct {
	config               config.Config
	externalConfig       config.ExternalConfig
	economicsConfig      config.EconomicsConfig
	shardCoordinator     sharding.Coordinator
	nodesCoordinator     nodesCoordinator.NodesCoordinator
	epochStartNotifier   factory.EpochStartNotifier
	forkDetector         process.ForkDetector
	coreComponents       factory.CoreComponentsHolder
	statusCoreComponents factory.StatusCoreComponentsHolder
	networkComponents    factory.NetworkComponentsHolder
	stateComponents      factory.StateComponentsHolder
	isInImportMode       bool
}

var log = logger.GetOrCreate("factory")

// NewStatusComponentsFactory will return a status components factory
func NewStatusComponentsFactory(args StatusComponentsFactoryArgs) (*statusComponentsFactory, error) {
	if check.IfNil(args.CoreComponents) {
		return nil, errors.ErrNilCoreComponentsHolder
	}
	if check.IfNil(args.CoreComponents.GenesisNodesSetup()) {
		return nil, errors.ErrNilGenesisNodesSetupHandler
	}
	if check.IfNil(args.NetworkComponents) {
		return nil, errors.ErrNilNetworkComponentsHolder
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, errors.ErrNilShardCoordinator
	}
	if check.IfNil(args.NodesCoordinator) {
		return nil, errors.ErrNilNodesCoordinator
	}
	if check.IfNil(args.EpochStartNotifier) {
		return nil, errors.ErrNilEpochStartNotifier
	}
	if check.IfNil(args.StatusCoreComponents) {
		return nil, errors.ErrNilStatusCoreComponents
	}

	return &statusComponentsFactory{
		config:               args.Config,
		externalConfig:       args.ExternalConfig,
		economicsConfig:      args.EconomicsConfig,
		shardCoordinator:     args.ShardCoordinator,
		nodesCoordinator:     args.NodesCoordinator,
		epochStartNotifier:   args.EpochStartNotifier,
		coreComponents:       args.CoreComponents,
		statusCoreComponents: args.StatusCoreComponents,
		networkComponents:    args.NetworkComponents,
		stateComponents:      args.StateComponents,
		isInImportMode:       args.IsInImportMode,
	}, nil
}

// Create will create and return the status components
func (scf *statusComponentsFactory) Create() (*statusComponents, error) {
	var err error

	log.Trace("creating software checker structure")
	softwareVersionCheckerFactory, err := swVersionFactory.NewSoftwareVersionFactory(
		scf.statusCoreComponents.AppStatusHandler(),
		scf.config.SoftwareVersionConfig,
	)
	if err != nil {
		return nil, err
	}

	softwareVersionChecker, err := softwareVersionCheckerFactory.Create()
	if err != nil {
		return nil, err
	}

	softwareVersionChecker.StartCheckSoftwareVersion()

	roundDurationSec := scf.coreComponents.GenesisNodesSetup().GetRoundDuration() / 1000
	if roundDurationSec < 1 {
		return nil, errors.ErrInvalidRoundDuration
	}

	outportHandler, err := scf.createOutportDriver()
	if err != nil {
		return nil, err
	}

	_, cancelFunc := context.WithCancel(context.Background())

	statusComponentsInstance := &statusComponents{
		nodesCoordinator: scf.nodesCoordinator,
		softwareVersion:  softwareVersionChecker,
		outportHandler:   outportHandler,
		statusHandler:    scf.statusCoreComponents.AppStatusHandler(),
		cancelFunc:       cancelFunc,
	}

	if scf.shardCoordinator.SelfId() == core.MetachainShardId {
		scf.epochStartNotifier.RegisterHandler(statusComponentsInstance.epochStartEventHandler())
	}

	return statusComponentsInstance, nil
}

func (pc *statusComponents) epochStartEventHandler() epochStart.ActionHandler {
	subscribeHandler := notifier.NewHandlerForEpochStart(func(hdr nodeData.HeaderHandler) {
		currentEpoch := hdr.GetEpoch()
		validatorsPubKeys, err := pc.nodesCoordinator.GetAllEligibleValidatorsPublicKeys(currentEpoch)
		if err != nil {
			log.Warn("pc.nodesCoordinator.GetAllEligibleValidatorPublicKeys for current epoch failed",
				"epoch", currentEpoch,
				"error", err.Error())
		}

		pc.outportHandler.SaveValidatorsPubKeys(&outportCore.ValidatorsPubKeys{
			ShardValidatorsPubKeys: outportCore.ConvertPubKeys(validatorsPubKeys),
			Epoch:                  currentEpoch,
		})

	}, func(_ nodeData.HeaderHandler) {}, common.IndexerOrder)

	return subscribeHandler
}

// IsInterfaceNil returns true if there is no value under the interface
func (scf *statusComponentsFactory) IsInterfaceNil() bool {
	return scf == nil
}

// Close closes all underlying components that need closing
func (pc *statusComponents) Close() error {
	pc.cancelFunc()

	if !check.IfNil(pc.softwareVersion) {
		log.LogIfError(pc.softwareVersion.Close())
	}

	return nil
}

// createOutportDriver creates a new outport.OutportHandler which is used to register outport drivers
// once a driver is subscribed it will receive data through the implemented outport.Driver methods
func (scf *statusComponentsFactory) createOutportDriver() (outport.OutportHandler, error) {
	hostDriverArgs, err := scf.makeHostDriverArgs()
	if err != nil {
		return nil, err
	}

	eventNotifierArgs, err := scf.makeEventNotifierArgs()
	if err != nil {
		return nil, err
	}

	outportFactoryArgs := &outportDriverFactory.OutportFactoryArgs{
		RetrialInterval:           common.RetrialIntervalForOutportDriver,
		ElasticIndexerFactoryArgs: scf.makeElasticIndexerArgs(),
		EventNotifierFactoryArgs:  eventNotifierArgs,
		HostDriverArgs:            hostDriverArgs,
	}

	return outportDriverFactory.CreateOutport(outportFactoryArgs)
}

func (scf *statusComponentsFactory) makeElasticIndexerArgs() indexerFactory.ArgsIndexerFactory {
	elasticSearchConfig := scf.externalConfig.ElasticSearchConnector
	return indexerFactory.ArgsIndexerFactory{
		Enabled:                  elasticSearchConfig.Enabled,
		IndexerCacheSize:         elasticSearchConfig.IndexerCacheSize,
		BulkRequestMaxSize:       elasticSearchConfig.BulkRequestMaxSizeInBytes,
		Url:                      elasticSearchConfig.URL,
		UserName:                 elasticSearchConfig.Username,
		Password:                 elasticSearchConfig.Password,
		Marshalizer:              scf.coreComponents.InternalMarshalizer(),
		Hasher:                   scf.coreComponents.Hasher(),
		AddressPubkeyConverter:   scf.coreComponents.AddressPubKeyConverter(),
		ValidatorPubkeyConverter: scf.coreComponents.ValidatorPubKeyConverter(),
		EnabledIndexes:           elasticSearchConfig.EnabledIndexes,
		Denomination:             scf.economicsConfig.GlobalSettings.Denomination,
		UseKibana:                elasticSearchConfig.UseKibana,
		ImportDB:                 scf.isInImportMode,
	}
}

func (scf *statusComponentsFactory) makeEventNotifierArgs() (*outportDriverFactory.EventNotifierFactoryArgs, error) {
	eventNotifierConfig := scf.externalConfig.EventNotifierConnector

	marshaller, err := factoryMarshalizer.NewMarshalizer(eventNotifierConfig.MarshallerType)
	if err != nil {
		return &outportDriverFactory.EventNotifierFactoryArgs{}, err
	}

	return &outportDriverFactory.EventNotifierFactoryArgs{
		Enabled:           eventNotifierConfig.Enabled,
		UseAuthorization:  eventNotifierConfig.UseAuthorization,
		ProxyUrl:          eventNotifierConfig.ProxyUrl,
		Username:          eventNotifierConfig.Username,
		Password:          eventNotifierConfig.Password,
		RequestTimeoutSec: eventNotifierConfig.RequestTimeoutSec,
		Marshaller:        marshaller,
	}, nil
}

func (scf *statusComponentsFactory) makeHostDriverArgs() (outportDriverFactory.ArgsHostDriverFactory, error) {
	if !scf.externalConfig.HostDriverConfig.Enabled {
		return outportDriverFactory.ArgsHostDriverFactory{}, nil
	}

	marshaller, err := factoryMarshalizer.NewMarshalizer(scf.externalConfig.HostDriverConfig.MarshallerType)
	if err != nil {
		return outportDriverFactory.ArgsHostDriverFactory{}, err
	}

	return outportDriverFactory.ArgsHostDriverFactory{
		Marshaller: marshaller,
		HostConfig: scf.externalConfig.HostDriverConfig,
	}, nil
}
