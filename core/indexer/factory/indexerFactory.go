package factory

import (
	"bytes"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/data/state"
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
	EnabledIndexes           []string
	AccountsDB               state.AccountsAdapter
}

// NewIndexer will create a new instance of Indexer
func NewIndexer(args *ArgsIndexerFactory) (indexer.Indexer, error) {
	err := checkDataIndexerParams(args)
	if err != nil {
		return nil, err
	}

	if !args.Enabled {
		return indexer.NewNilIndexer(), nil
	}

	elasticProcessor, err := createElasticProcessor(args)
	if err != nil {
		return nil, err
	}

	dispatcher, err := indexer.NewDataDispatcher(args.IndexerCacheSize)
	if err != nil {
		return nil, err
	}

	dispatcher.StartIndexData()

	arguments := indexer.ArgDataIndexer{
		Marshalizer:        args.Marshalizer,
		Options:            args.Options,
		NodesCoordinator:   args.NodesCoordinator,
		EpochStartNotifier: args.EpochStartNotifier,
		ShardID:            args.ShardID,
		ElasticProcessor:   elasticProcessor,
		DataDispatcher:     dispatcher,
	}

	return indexer.NewDataIndexer(arguments)
}

func createDatabaseClient(url, userName, password string) (indexer.DatabaseClientHandler, error) {
	return indexer.NewElasticClient(elasticsearch.Config{
		Addresses: []string{url},
		Username:  userName,
		Password:  password,
	})
}

func createElasticProcessor(args *ArgsIndexerFactory) (indexer.ElasticProcessor, error) {
	databaseClient, err := createDatabaseClient(args.Url, args.UserName, args.Password)
	if err != nil {
		return nil, err
	}

	enabledIndexesMap := make(map[string]struct{})
	for _, index := range args.EnabledIndexes {
		enabledIndexesMap[index] = struct{}{}
	}
	if len(enabledIndexesMap) == 0 {
		return nil, fmt.Errorf("empty elastic search enabled indexes map")
	}

	esIndexerArgs := indexer.ArgElasticProcessor{
		IndexTemplates:           args.IndexTemplates,
		IndexPolicies:            args.IndexPolicies,
		Marshalizer:              args.Marshalizer,
		Hasher:                   args.Hasher,
		AddressPubkeyConverter:   args.AddressPubkeyConverter,
		ValidatorPubkeyConverter: args.ValidatorPubkeyConverter,
		Options:                  args.Options,
		DBClient:                 databaseClient,
		EnabledIndexes:           enabledIndexesMap,
		AccountsDB:               args.AccountsDB,
	}

	return indexer.NewElasticProcessor(esIndexerArgs)
}

func checkDataIndexerParams(arguments *ArgsIndexerFactory) error {
	if check.IfNil(arguments.AddressPubkeyConverter) {
		return fmt.Errorf("%w when setting AddressPubkeyConverter in indexer", indexer.ErrNilPubkeyConverter)
	}
	if check.IfNil(arguments.ValidatorPubkeyConverter) {
		return fmt.Errorf("%w when setting ValidatorPubkeyConverter in indexer", indexer.ErrNilPubkeyConverter)
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
	if check.IfNil(arguments.EpochStartNotifier) {
		return core.ErrNilEpochStartNotifier
	}

	return nil
}
