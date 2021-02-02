package factory

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/core/indexer/disabled"
	processIndexer "github.com/ElrondNetwork/elrond-go/core/indexer/process"
	"github.com/ElrondNetwork/elrond-go/core/indexer/process/accounts"
	blockProc "github.com/ElrondNetwork/elrond-go/core/indexer/process/block"
	"github.com/ElrondNetwork/elrond-go/core/indexer/process/generalInfo"
	"github.com/ElrondNetwork/elrond-go/core/indexer/process/miniblocks"
	"github.com/ElrondNetwork/elrond-go/core/indexer/process/transactions"
	"github.com/ElrondNetwork/elrond-go/core/indexer/process/validators"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
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
	templatesAndPoliciesReader := processIndexer.NewTemplatesAndPoliciesReader(arguments.TemplatesPath, arguments.UseKibana)
	indexTemplates, indexPolicies, err := templatesAndPoliciesReader.GetElasticTemplatesAndPolicies()
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
	miniblocksProc := miniblocks.NewMiniblocksProcessor(arguments.ShardCoordinator.SelfId(), arguments.Hasher, arguments.Marshalizer)
	validatorsProc := validators.NewValidatorsProcessor(arguments.ValidatorPubkeyConverter)
	generalInfoProc := generalInfo.NewGeneralInfoProcessor()

	txsProc := transactions.NewTransactionsProcessor(
		arguments.AddressPubkeyConverter,
		arguments.TransactionFeeCalculator,
		arguments.IsInImportDBMode,
		arguments.ShardCoordinator,
		arguments.SaveTxsLogsEnabled,
		disabled.NewNilTxLogsProcessor(),
		arguments.Hasher,
		arguments.Marshalizer,
	)

	args := &processIndexer.ArgElasticProcessor{
		TxProc:          txsProc,
		AccountsProc:    accountsProc,
		BlockProc:       blockProcHandler,
		MiniblocksProc:  miniblocksProc,
		ValidatorsProc:  validatorsProc,
		GeneralInfoProc: generalInfoProc,
		DBClient:        arguments.DBClient,
		EnabledIndexes:  enabledIndexesMap,
		UseKibana:       arguments.UseKibana,
		IndexTemplates:  indexTemplates,
		IndexPolicies:   indexPolicies,
		SelfShardID:     arguments.ShardCoordinator.SelfId(),
	}

	return processIndexer.NewElasticProcessor(args)
}
