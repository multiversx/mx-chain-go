package factory

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/elastic/go-elasticsearch/v7"
)

// ArgsIndexerFactory holds all dependencies required by the data indexer factory in order to create
// new instances
type ArgsIndexerFactory struct {
	Enabled                  bool
	IndexerCacheSize         int
	ShardID                  uint32
	Url                      string
	UserName                 string
	Password                 string
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	EpochStartNotifier       sharding.EpochStartEventNotifier
	NodesCoordinator         sharding.NodesCoordinator
	AddressPubkeyConverter   core.PubkeyConverter
	ValidatorPubkeyConverter core.PubkeyConverter
	IndexTemplates           map[string]*bytes.Buffer
	IndexPolicies            map[string]*bytes.Buffer
	Options                  *indexer.Options
}

type dataIndexerFactory struct {
	enabled                  bool
	indexerCacheSize         int
	shardID                  uint32
	url                      string
	userName                 string
	password                 string
	marshalizer              marshal.Marshalizer
	hasher                   hashing.Hasher
	epochStartNotifier       sharding.EpochStartEventNotifier
	nodesCoordinator         sharding.NodesCoordinator
	addressPubkeyConverter   core.PubkeyConverter
	validatorPubkeyConverter core.PubkeyConverter
	indexTemplates           map[string]*bytes.Buffer
	indexPolicies            map[string]*bytes.Buffer
	options                  *indexer.Options
}

// NewIndexerFactory will create a new instance of dataIndexerFactory
func NewIndexerFactory(args *ArgsIndexerFactory) (indexer.DataIndexerFactory, error) {
	err := checkDataIndexerParams(args)
	if err != nil {
		return nil, err
	}

	return &dataIndexerFactory{
		enabled:                  args.Enabled,
		indexerCacheSize:         args.IndexerCacheSize,
		shardID:                  args.ShardID,
		url:                      args.Url,
		userName:                 args.UserName,
		password:                 args.Password,
		marshalizer:              args.Marshalizer,
		hasher:                   args.Hasher,
		epochStartNotifier:       args.EpochStartNotifier,
		nodesCoordinator:         args.NodesCoordinator,
		addressPubkeyConverter:   args.AddressPubkeyConverter,
		validatorPubkeyConverter: args.ValidatorPubkeyConverter,
		indexTemplates:           args.IndexTemplates,
		indexPolicies:            args.IndexPolicies,
		options:                  args.Options,
	}, nil
}

// Create creates instances of HistoryRepository
func (dif *dataIndexerFactory) Create() (indexer.Indexer, error) {
	if !dif.enabled {
		return indexer.NewNilIndexer(), nil
	}

	elasticProcessor, err := dif.createElasticProcessor()
	if err != nil {
		return nil, err
	}

	dispatcher, err := indexer.NewDataDispatcher(dif.indexerCacheSize)
	if err != nil {
		return nil, err
	}

	dispatcher.StartIndexData()

	arguments := indexer.DataIndexerArgs{
		TxIndexingEnabled:  dif.options.TxIndexingEnabled,
		Marshalizer:        dif.marshalizer,
		Options:            dif.options,
		NodesCoordinator:   dif.nodesCoordinator,
		EpochStartNotifier: dif.epochStartNotifier,
		ShardID:            dif.shardID,
		ElasticProcessor:   elasticProcessor,
		DataDispatcher:     dispatcher,
	}

	return indexer.NewDataIndexer(arguments)
}

func (dif *dataIndexerFactory) createDatabaseClient() (indexer.DatabaseClientHandler, error) {
	return indexer.NewElasticClient(elasticsearch.Config{
		Addresses: []string{dif.url},
		Username:  dif.userName,
		Password:  dif.password,
	})
}

func (dif *dataIndexerFactory) createElasticProcessor() (indexer.ElasticProcessor, error) {
	databaseClient, err := dif.createDatabaseClient()
	if err != nil {
		return nil, err
	}

	esIndexerArgs := indexer.ElasticProcessorArgs{
		IndexTemplates:           dif.indexTemplates,
		IndexPolicies:            dif.indexPolicies,
		Marshalizer:              dif.marshalizer,
		Hasher:                   dif.hasher,
		AddressPubkeyConverter:   dif.addressPubkeyConverter,
		ValidatorPubkeyConverter: dif.validatorPubkeyConverter,
		Options:                  dif.options,
		DBClient:                 databaseClient,
	}

	return indexer.NewElasticProcessor(esIndexerArgs)
}

// IsInterfaceNil returns true if there is no value under the interface
func (dif *dataIndexerFactory) IsInterfaceNil() bool {
	return dif == nil
}

func checkDataIndexerParams(arguments *ArgsIndexerFactory) error {
	if arguments.IndexerCacheSize <= 0 {
		return indexer.ErrInvalidCacheSize
	}
	if check.IfNil(arguments.AddressPubkeyConverter) {
		return indexer.ErrNilPubkeyConverter
	}
	if check.IfNil(arguments.ValidatorPubkeyConverter) {
		return indexer.ErrNilPubkeyConverter
	}
	if arguments.Url == "" {
		return core.ErrNilUrl
	}
	if check.IfNil(arguments.Marshalizer) {
		return core.ErrNilMarshalizer
	}
	if check.IfNil(arguments.Hasher) {
		return core.ErrNilHasher
	}
	if check.IfNil(arguments.NodesCoordinator) {
		return core.ErrNilNodesCoordinator
	}
	if arguments.EpochStartNotifier == nil {
		return core.ErrNilEpochStartNotifier
	}

	return nil
}
