package factory

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/core/indexer/client"
	processIndexer "github.com/ElrondNetwork/elrond-go/core/indexer/process"
	"github.com/ElrondNetwork/elrond-go/core/indexer/process/factory"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/elastic/go-elasticsearch/v7"
)

// ArgsIndexerFactory holds all dependencies required by the data indexer factory in order to create
// new instances
type ArgsIndexerFactory struct {
	Enabled                  bool
	UseKibana                bool
	IsInImportDBMode         bool
	SaveTxsLogsEnabled       bool
	IndexerCacheSize         int
	Denomination             int
	Url                      string
	UserName                 string
	Password                 string
	TemplatesPath            string
	EnabledIndexes           []string
	ShardCoordinator         sharding.Coordinator
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	EpochStartNotifier       sharding.EpochStartEventNotifier
	NodesCoordinator         sharding.NodesCoordinator
	AddressPubkeyConverter   core.PubkeyConverter
	ValidatorPubkeyConverter core.PubkeyConverter
	AccountsDB               state.AccountsAdapter
	TransactionFeeCalculator process.TransactionFeeCalculator
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
		NodesCoordinator:   args.NodesCoordinator,
		EpochStartNotifier: args.EpochStartNotifier,
		ShardCoordinator:   args.ShardCoordinator,
		ElasticProcessor:   elasticProcessor,
		DataDispatcher:     dispatcher,
	}

	return indexer.NewDataIndexer(arguments)
}

func createElasticProcessor(args *ArgsIndexerFactory) (indexer.ElasticProcessor, error) {
	databaseClient, err := client.NewElasticClient(elasticsearch.Config{
		Addresses: []string{args.Url},
		Username:  args.UserName,
		Password:  args.Password,
	})
	if err != nil {
		return nil, err
	}

	argsElasticProcFac := factory.ArgElasticProcessorFactory{
		Marshalizer:              args.Marshalizer,
		Hasher:                   args.Hasher,
		AddressPubkeyConverter:   args.AddressPubkeyConverter,
		ValidatorPubkeyConverter: args.ValidatorPubkeyConverter,
		UseKibana:                args.UseKibana,
		DBClient:                 databaseClient,
		AccountsDB:               args.AccountsDB,
		Denomination:             args.Denomination,
		TransactionFeeCalculator: args.TransactionFeeCalculator,
		IsInImportDBMode:         args.IsInImportDBMode,
		ShardCoordinator:         args.ShardCoordinator,
		SaveTxsLogsEnabled:       args.SaveTxsLogsEnabled,
		EnabledIndexes:           args.EnabledIndexes,
		TemplatesPath:            args.TemplatesPath,
	}

	return factory.CreateElasticProcessor(argsElasticProcFac)
}

func checkDataIndexerParams(arguments *ArgsIndexerFactory) error {
	if arguments.IndexerCacheSize < 0 {
		return indexer.ErrNegativeCacheSize
	}
	if check.IfNil(arguments.AddressPubkeyConverter) {
		return fmt.Errorf("%w when setting AddressPubkeyConverter in indexer", processIndexer.ErrNilPubkeyConverter)
	}
	if check.IfNil(arguments.ValidatorPubkeyConverter) {
		return fmt.Errorf("%w when setting ValidatorPubkeyConverter in indexer", processIndexer.ErrNilPubkeyConverter)
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
	if check.IfNil(arguments.TransactionFeeCalculator) {
		return core.ErrNilTransactionFeeCalculator
	}
	if check.IfNil(arguments.AccountsDB) {
		return processIndexer.ErrNilAccountsDB
	}
	if check.IfNil(arguments.ShardCoordinator) {
		return processIndexer.ErrNilShardCoordinator
	}

	return nil
}
