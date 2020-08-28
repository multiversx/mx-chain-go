package factory

import (
	"context"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/core/statistics/softwareVersion/factory"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// TODO: move app status handler initialization here

type statusComponents struct {
	statusHandler   core.AppStatusHandler
	tpsBenchmark    statistics.TPSBenchmark
	elasticIndexer  indexer.Indexer
	softwareVersion statistics.SoftwareVersionChecker
	cancelFunc      func()
}

// StatusComponentsFactoryArgs redefines the arguments structure needed for the status components factory
type StatusComponentsFactoryArgs struct {
	Config             config.Config
	ExternalConfig     config.ExternalConfig
	RoundDurationSec   uint64
	ElasticOptions     *indexer.Options
	StatusUtils        StatusHandlersUtils
	ShardCoordinator   sharding.Coordinator
	NodesCoordinator   sharding.NodesCoordinator
	EpochStartNotifier EpochStartNotifier
	CoreComponents     CoreComponentsHolder
	DataComponents     DataComponentsHolder
	NetworkComponents  NetworkComponentsHolder
}

type statusComponentsFactory struct {
	config              config.Config
	externalConfig      config.ExternalConfig
	roundDurationSec    uint64
	elasticOptions      *indexer.Options
	statusHandlersUtils StatusHandlersUtils
	shardCoordinator    sharding.Coordinator
	nodesCoordinator    sharding.NodesCoordinator
	epochStartNotifier  EpochStartNotifier
	forkDetector        process.ForkDetector
	coreComponents      CoreComponentsHolder
	dataComponents      DataComponentsHolder
	networkComponents   NetworkComponentsHolder
}

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
	if args.RoundDurationSec < 1 {
		return nil, errors.ErrInvalidRoundDuration
	}
	if check.IfNil(args.StatusUtils) {
		return nil, errors.ErrNilStatusHandlersUtils
	}

	if args.ElasticOptions == nil {
		return nil, errors.ErrNilElasticOptions
	}

	return &statusComponentsFactory{
		config:              args.Config,
		externalConfig:      args.ExternalConfig,
		roundDurationSec:    args.RoundDurationSec,
		elasticOptions:      args.ElasticOptions,
		shardCoordinator:    args.ShardCoordinator,
		nodesCoordinator:    args.NodesCoordinator,
		epochStartNotifier:  args.EpochStartNotifier,
		coreComponents:      args.CoreComponents,
		dataComponents:      args.DataComponents,
		networkComponents:   args.NetworkComponents,
		statusHandlersUtils: args.StatusUtils,
	}, nil
}

// Create will create and return the status components
func (scf *statusComponentsFactory) Create() (*statusComponents, error) {
	_, cancelFunc := context.WithCancel(context.Background())

	softwareVersionCheckerFactory, err := factory.NewSoftwareVersionFactory(
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

	initialTpsBenchmark := scf.statusHandlersUtils.LoadTpsBenchmarkFromStorage(
		scf.dataComponents.StorageService().GetStorer(dataRetriever.StatusMetricsUnit),
		scf.coreComponents.InternalMarshalizer(),
	)

	tpsBenchmark, err := statistics.NewTPSBenchmarkWithInitialData(
		scf.coreComponents.StatusHandler(),
		initialTpsBenchmark,
		scf.shardCoordinator.NumberOfShards(),
		scf.roundDurationSec,
	)
	if err != nil {
		return nil, err
	}

	elasticIndexer, err := scf.createElasticIndexer()
	if err != nil {
		return nil, err
	}

	return &statusComponents{
		softwareVersion: softwareVersionChecker,
		tpsBenchmark:    tpsBenchmark,
		elasticIndexer:  elasticIndexer,
		statusHandler:   scf.coreComponents.StatusHandler(),
		cancelFunc:      cancelFunc,
	}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (scf *statusComponentsFactory) IsInterfaceNil() bool {
	return scf == nil
}

// Closes all underlying components that need closing
func (pc *statusComponents) Close() error {
	pc.cancelFunc()

	// TODO: close all components

	return nil
}

// createElasticIndexer creates a new elasticIndexer where the server listens on the url,
// authentication for the server is using the username and password
func (scf *statusComponentsFactory) createElasticIndexer() (indexer.Indexer, error) {
	if !scf.externalConfig.ElasticSearchConnector.Enabled {
		log.Debug("elastic search indexing not enabled, will create a NilIndexer")
		return indexer.NewNilIndexer(), nil
	}

	log.Debug("elastic search indexing enabled, will create an ElasticIndexer")
	elasticIndexerArgs := indexer.ElasticIndexerArgs{
		ShardId:                  scf.shardCoordinator.SelfId(),
		Url:                      scf.externalConfig.ElasticSearchConnector.URL,
		UserName:                 scf.externalConfig.ElasticSearchConnector.Username,
		Password:                 scf.externalConfig.ElasticSearchConnector.Password,
		Marshalizer:              scf.coreComponents.VmMarshalizer(),
		Hasher:                   scf.coreComponents.Hasher(),
		EpochStartNotifier:       scf.epochStartNotifier,
		NodesCoordinator:         scf.nodesCoordinator,
		AddressPubkeyConverter:   scf.coreComponents.AddressPubKeyConverter(),
		ValidatorPubkeyConverter: scf.coreComponents.ValidatorPubKeyConverter(),
		Options:                  scf.elasticOptions,
	}

	return indexer.NewElasticIndexer(elasticIndexerArgs)
}
