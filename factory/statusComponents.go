package factory

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/core/statistics/softwareVersion/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// TODO: move app status handler initialization here

type statusComponents struct {
	statusHandler   core.AppStatusHandler
	tpsBenchmark    statistics.TPSBenchmark
	elasticSearch   indexer.Indexer
	softwareVersion statistics.SoftwareVersionChecker
}

// StatusComponentsFactoryArgs redefines the arguments structure needed for the status components factory
type StatusComponentsFactoryArgs struct {
	Config            config.Config
	ExternalConfig    config.ExternalConfig
	RoundDurationSec  uint64
	ElasticOptions    *indexer.Options
	ShardCoordinator  sharding.Coordinator
	CoreComponents    CoreComponentsHolder
	NetworkComponents NetworkComponentsHolder
	ProcessComponents ProcessComponentsHolder
}

type statusComponentsFactory struct {
	config            config.Config
	externalConfig    config.ExternalConfig
	roundDuration     uint64
	elasticOptions    *indexer.Options
	shardCoordinator  sharding.Coordinator
	coreComponents    CoreComponentsHolder
	networkComponents NetworkComponentsHolder
	processComponents ProcessComponentsHolder
}

// NewStatusComponentsFactory will return a status components factory
func NewStatusComponentsFactory(args StatusComponentsFactoryArgs) (*statusComponentsFactory, error) {
	if check.IfNil(args.CoreComponents) {
		return nil, ErrNilCoreComponentsHolder
	}
	if check.IfNil(args.NetworkComponents) {
		return nil, ErrNilNetworkComponentsHolder
	}
	if check.IfNil(args.ProcessComponents) {
		return nil, ErrNilProcessComponentsHolder
	}
	if check.IfNil(args.CoreComponents.AddressPubKeyConverter()) {
		return nil, fmt.Errorf("%w for address", ErrNilPubKeyConverter)
	}
	if check.IfNil(args.CoreComponents.ValidatorPubKeyConverter()) {
		return nil, fmt.Errorf("%w for validator", ErrNilPubKeyConverter)
	}
	if check.IfNil(args.ProcessComponents.NodesCoordinator()) {
		return nil, ErrNilNodesCoordinator
	}
	if args.RoundDurationSec < 1 {
		return nil, ErrInvalidRoundDuration
	}
	if check.IfNil(args.ProcessComponents.EpochStartNotifier()) {
		return nil, ErrNilEpochStartNotifier
	}
	if args.ElasticOptions == nil {
		return nil, ErrNilElasticOptions
	}

	return &statusComponentsFactory{
		config:            args.Config,
		externalConfig:    args.ExternalConfig,
		roundDuration:     args.RoundDurationSec,
		elasticOptions:    args.ElasticOptions,
		shardCoordinator:  args.ShardCoordinator,
		coreComponents:    args.CoreComponents,
		networkComponents: args.NetworkComponents,
		processComponents: args.ProcessComponents,
	}, nil
}

// Create will create and return the status components
func (scf *statusComponentsFactory) Create() (*statusComponents, error) {
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

	tpsBenchmark, err := statistics.NewTPSBenchmark(scf.shardCoordinator.NumberOfShards(), scf.roundDuration)
	if err != nil {
		return nil, err
	}

	var elasticIndexer indexer.Indexer

	if scf.externalConfig.ElasticSearchConnector.Enabled {
		elasticIndexerArgs := indexer.ElasticIndexerArgs{
			ShardId:                  scf.shardCoordinator.SelfId(),
			Url:                      scf.externalConfig.ElasticSearchConnector.URL,
			UserName:                 scf.externalConfig.ElasticSearchConnector.Username,
			Password:                 scf.externalConfig.ElasticSearchConnector.Password,
			Marshalizer:              scf.coreComponents.VmMarshalizer(),
			Hasher:                   scf.coreComponents.Hasher(),
			EpochStartNotifier:       scf.processComponents.EpochStartNotifier(),
			NodesCoordinator:         scf.processComponents.NodesCoordinator(),
			AddressPubkeyConverter:   scf.coreComponents.AddressPubKeyConverter(),
			ValidatorPubkeyConverter: scf.coreComponents.ValidatorPubKeyConverter(),
			Options:                  scf.elasticOptions,
		}
		elasticIndexer, err = indexer.NewElasticIndexer(elasticIndexerArgs)
		if err != nil {
			return nil, err
		}
	} else {
		elasticIndexer = indexer.NewNilIndexer()
	}

	return &statusComponents{
		softwareVersion: softwareVersionChecker,
		tpsBenchmark:    tpsBenchmark,
		elasticSearch:   elasticIndexer,
	}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (scf *statusComponentsFactory) IsInterfaceNil() bool {
	return scf == nil
}
