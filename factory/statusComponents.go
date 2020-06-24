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

type statusComponentsFactory struct {
	shardCoordinator         sharding.Coordinator
	softwareVersionConfig    config.SoftwareVersionConfig
	elasticConfig            config.ElasticSearchConfig
	coreData                 CoreComponentsHolder
	epochStartNotifier       EpochStartNotifier
	nodesCoordinator         sharding.NodesCoordinator
	addressPubkeyConverter   core.PubkeyConverter
	validatorPubkeyConverter core.PubkeyConverter
	elasticOptions           *indexer.Options
	roundDurationSec         int
}

// StatusComponentsFactoryArgs holds the arguments needed for creating a status components factory
type StatusComponentsFactoryArgs struct {
	CoreData                 CoreComponentsHolder
	ShardCoordinator         sharding.Coordinator
	SoftwareVersionConfig    config.SoftwareVersionConfig
	ElasticConfig            config.ElasticSearchConfig
	EpochNotifier            EpochStartNotifier
	NodesCoordinator         sharding.NodesCoordinator
	AddressPubkeyConverter   core.PubkeyConverter
	ValidatorPubkeyConverter core.PubkeyConverter
	ElasticOptions           *indexer.Options
	RoundDurationSec         int
}

// NewStatusComponentsFactory will return a status components factory
func NewStatusComponentsFactory(args StatusComponentsFactoryArgs) (*statusComponentsFactory, error) {
	if args.CoreData == nil {
		return nil, ErrNilCoreComponentsHolder
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, ErrNilShardCoordinator
	}
	if args.RoundDurationSec < 1 {
		return nil, ErrInvalidRoundDuration
	}
	if check.IfNil(args.EpochNotifier) {
		return nil, ErrNilEpochStartNotifier
	}
	if check.IfNil(args.NodesCoordinator) {
		return nil, ErrNilNodesCoordinator
	}
	if check.IfNil(args.AddressPubkeyConverter) {
		return nil, fmt.Errorf("%w for address", ErrNilPubKeyConverter)
	}
	if check.IfNil(args.ValidatorPubkeyConverter) {
		return nil, fmt.Errorf("%w for validator", ErrNilPubKeyConverter)
	}
	if args.ElasticOptions == nil {
		return nil, ErrNilElasticOptions
	}

	return &statusComponentsFactory{
		shardCoordinator:         args.ShardCoordinator,
		roundDurationSec:         args.RoundDurationSec,
		elasticConfig:            args.ElasticConfig,
		coreData:                 args.CoreData,
		epochStartNotifier:       args.EpochNotifier,
		nodesCoordinator:         args.NodesCoordinator,
		addressPubkeyConverter:   args.AddressPubkeyConverter,
		validatorPubkeyConverter: args.ValidatorPubkeyConverter,
		elasticOptions:           args.ElasticOptions,
		softwareVersionConfig:    args.SoftwareVersionConfig,
	}, nil
}

// Create will create and return the status components
func (scf *statusComponentsFactory) Create() (*statusComponents, error) {
	softwareVersionCheckerFactory, err := factory.NewSoftwareVersionFactory(scf.coreData.StatusHandler(), scf.softwareVersionConfig)
	if err != nil {
		return nil, err
	}
	softwareVersionChecker, err := softwareVersionCheckerFactory.Create()
	if err != nil {
		return nil, err
	}

	tpsBenchmark, err := statistics.NewTPSBenchmark(scf.shardCoordinator.NumberOfShards(), uint64(scf.roundDurationSec))
	if err != nil {
		return nil, err
	}

	var elasticIndexer indexer.Indexer
	if scf.elasticConfig.Enabled {
		elasticIndexerArgs := indexer.ElasticIndexerArgs{
			ShardId:                  scf.shardCoordinator.SelfId(),
			Url:                      scf.elasticConfig.URL,
			UserName:                 scf.elasticConfig.Username,
			Password:                 scf.elasticConfig.Password,
			Marshalizer:              scf.coreData.VmMarshalizer(),
			Hasher:                   scf.coreData.Hasher(),
			EpochStartNotifier:       scf.epochStartNotifier,
			NodesCoordinator:         scf.nodesCoordinator,
			AddressPubkeyConverter:   scf.addressPubkeyConverter,
			ValidatorPubkeyConverter: scf.validatorPubkeyConverter,
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
