package status

import (
	"context"
	"fmt"

	covalentFactory "github.com/ElrondNetwork/covalent-indexer-go/factory"
	indexerFactory "github.com/ElrondNetwork/elastic-indexer-go/factory"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	nodeData "github.com/ElrondNetwork/elrond-go-core/data"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/common/statistics"
	swVersionFactory "github.com/ElrondNetwork/elrond-go/common/statistics/softwareVersion/factory"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/outport"
	outportDriverFactory "github.com/ElrondNetwork/elrond-go/outport/factory"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
)

// TODO: move app status handler initialization here

type statusComponents struct {
	nodesCoordinator nodesCoordinator.NodesCoordinator
	statusHandler    core.AppStatusHandler
	outportHandler   outport.OutportHandler
	softwareVersion  statistics.SoftwareVersionChecker
	resourceMonitor  statistics.ResourceMonitorHandler
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
	DataComponents       factory.DataComponentsHolder
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
	dataComponents       factory.DataComponentsHolder
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
	if check.IfNil(args.DataComponents) {
		return nil, errors.ErrNilDataComponentsHolder
	}
	if check.IfNil(args.NetworkComponents) {
		return nil, errors.ErrNilNetworkComponentsHolder
	}
	if check.IfNil(args.CoreComponents.AddressPubKeyConverter()) {
		return nil, fmt.Errorf("%w for address", errors.ErrNilPubKeyConverter)
	}
	if check.IfNil(args.CoreComponents.ValidatorPubKeyConverter()) {
		return nil, fmt.Errorf("%w for validator", errors.ErrNilPubKeyConverter)
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
		dataComponents:       args.DataComponents,
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
		scf.coreComponents.StatusHandler(),
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
		statusHandler:    scf.coreComponents.StatusHandler(),
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

		pc.outportHandler.SaveValidatorsPubKeys(validatorsPubKeys, currentEpoch)

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

	outportFactoryArgs := &outportDriverFactory.OutportFactoryArgs{
		RetrialInterval:            common.RetrialIntervalForOutportDriver,
		ElasticIndexerFactoryArgs:  scf.makeElasticIndexerArgs(),
		EventNotifierFactoryArgs:   scf.makeEventNotifierArgs(),
		CovalentIndexerFactoryArgs: scf.makeCovalentIndexerArgs(),
	}

	return outportDriverFactory.CreateOutport(outportFactoryArgs)
}

func (scf *statusComponentsFactory) makeElasticIndexerArgs() *indexerFactory.ArgsIndexerFactory {
	elasticSearchConfig := scf.externalConfig.ElasticSearchConnector
	return &indexerFactory.ArgsIndexerFactory{
		Enabled:                  elasticSearchConfig.Enabled,
		IndexerCacheSize:         elasticSearchConfig.IndexerCacheSize,
		BulkRequestMaxSize:       elasticSearchConfig.BulkRequestMaxSizeInBytes,
		ShardCoordinator:         scf.shardCoordinator,
		Url:                      elasticSearchConfig.URL,
		UserName:                 elasticSearchConfig.Username,
		Password:                 elasticSearchConfig.Password,
		Marshalizer:              scf.coreComponents.InternalMarshalizer(),
		Hasher:                   scf.coreComponents.Hasher(),
		AddressPubkeyConverter:   scf.coreComponents.AddressPubKeyConverter(),
		ValidatorPubkeyConverter: scf.coreComponents.ValidatorPubKeyConverter(),
		EnabledIndexes:           elasticSearchConfig.EnabledIndexes,
		AccountsDB:               scf.stateComponents.AccountsAdapter(),
		Denomination:             scf.economicsConfig.GlobalSettings.Denomination,
		TransactionFeeCalculator: scf.coreComponents.EconomicsData(),
		UseKibana:                elasticSearchConfig.UseKibana,
		IsInImportDBMode:         scf.isInImportMode,
	}
}

func (scf *statusComponentsFactory) makeEventNotifierArgs() *outportDriverFactory.EventNotifierFactoryArgs {
	eventNotifierConfig := scf.externalConfig.EventNotifierConnector
	return &outportDriverFactory.EventNotifierFactoryArgs{
		Enabled:           eventNotifierConfig.Enabled,
		UseAuthorization:  eventNotifierConfig.UseAuthorization,
		ProxyUrl:          eventNotifierConfig.ProxyUrl,
		Username:          eventNotifierConfig.Username,
		Password:          eventNotifierConfig.Password,
		RequestTimeoutSec: eventNotifierConfig.RequestTimeoutSec,
		Marshaller:        scf.coreComponents.InternalMarshalizer(),
		Hasher:            scf.coreComponents.Hasher(),
		PubKeyConverter:   scf.coreComponents.AddressPubKeyConverter(),
	}
}

func (scf *statusComponentsFactory) makeCovalentIndexerArgs() *covalentFactory.ArgsCovalentIndexerFactory {
	return &covalentFactory.ArgsCovalentIndexerFactory{
		Enabled:              scf.externalConfig.CovalentConnector.Enabled,
		URL:                  scf.externalConfig.CovalentConnector.URL,
		RouteSendData:        scf.externalConfig.CovalentConnector.RouteSendData,
		RouteAcknowledgeData: scf.externalConfig.CovalentConnector.RouteAcknowledgeData,
		PubKeyConverter:      scf.coreComponents.AddressPubKeyConverter(),
		Accounts:             scf.stateComponents.AccountsAdapter(),
		Hasher:               scf.coreComponents.Hasher(),
		Marshaller:           scf.coreComponents.InternalMarshalizer(),
		ShardCoordinator:     scf.shardCoordinator,
	}
}
