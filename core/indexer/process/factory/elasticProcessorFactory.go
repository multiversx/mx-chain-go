package factory

import (
	"path"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/core/indexer/disabled"
	processIndexer "github.com/ElrondNetwork/elrond-go/core/indexer/process"
	"github.com/ElrondNetwork/elrond-go/core/indexer/process/accounts"
	blockProc "github.com/ElrondNetwork/elrond-go/core/indexer/process/block"
	"github.com/ElrondNetwork/elrond-go/core/indexer/process/miniblocks"
	"github.com/ElrondNetwork/elrond-go/core/indexer/process/transactions"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

const (
	withKibanaFolder = "withKibana"
	noKibanaFolder   = "noKibana"
)

//ArgElasticProcessorFactory is struct that is used to store all components that are needed to an elastic indexer
type ArgElasticProcessorFactory struct {
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	AddressPubkeyConverter   core.PubkeyConverter
	ValidatorPubkeyConverter core.PubkeyConverter
	DBClient                 processIndexer.DatabaseClientHandler
	AccountsDB               state.AccountsAdapter
	ShardCoordinator         sharding.Coordinator
	TransactionFeeCalculator process.TransactionFeeCalculator
	EnabledIndexes           []string
	TemplatesPath            string
	Denomination             int
	SaveTxsLogsEnabled       bool
	IsInImportDBMode         bool
	UseKibana                bool
}

// CreateElasticProcessor will create a new instance of ElasticProcessor
func CreateElasticProcessor(arguments ArgElasticProcessorFactory) (indexer.ElasticProcessor, error) {
	var templatesPath string
	if arguments.UseKibana {
		templatesPath = path.Join(arguments.TemplatesPath, withKibanaFolder)
	} else {
		templatesPath = path.Join(arguments.TemplatesPath, noKibanaFolder)
	}

	indexTemplates, indexPolicies, err := processIndexer.GetElasticTemplatesAndPolicies(templatesPath, arguments.UseKibana)
	if err != nil {
		return nil, err
	}

	enabledIndexesMap := make(map[string]struct{})
	for _, index := range arguments.EnabledIndexes {
		enabledIndexesMap[index] = struct{}{}
	}
	if len(enabledIndexesMap) == 0 {
		return nil, indexer.ErrEmptyEnabledIndexes
	}

	accountsProc := accounts.NewAccountsProcessor(arguments.Denomination, arguments.Marshalizer, arguments.AddressPubkeyConverter, arguments.AccountsDB)
	blockProcHandler := blockProc.NewBlockProcessor(arguments.Hasher, arguments.Marshalizer)
	miniblocksProc := miniblocks.NewMiniblocksProcessor(arguments.Hasher, arguments.Marshalizer)

	calculateHashFunc := func(object interface{}) ([]byte, error) {
		return core.CalculateHash(arguments.Marshalizer, arguments.Hasher, object)
	}
	txsProc := transactions.NewTransactionsProcessor(
		arguments.AddressPubkeyConverter,
		arguments.TransactionFeeCalculator,
		arguments.IsInImportDBMode,
		arguments.ShardCoordinator,
		arguments.SaveTxsLogsEnabled,
		calculateHashFunc,
		disabled.NewNilTxLogsProcessor(),
	)

	args := &processIndexer.ArgElasticProcessor{
		TxProc:                   txsProc,
		AccountsProc:             accountsProc,
		BlockProc:                blockProcHandler,
		MiniblocksProc:           miniblocksProc,
		DBClient:                 arguments.DBClient,
		EnabledIndexes:           enabledIndexesMap,
		CalculateHashFunc:        calculateHashFunc,
		ValidatorPubkeyConverter: arguments.ValidatorPubkeyConverter,
		UseKibana:                arguments.UseKibana,
		IndexTemplates:           indexTemplates,
		IndexPolicies:            indexPolicies,
		SelfShardID:              arguments.ShardCoordinator.SelfId(),
	}

	return processIndexer.NewElasticProcessor(args)
}
