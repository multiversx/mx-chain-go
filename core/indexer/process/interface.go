package process

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

// DatabaseClientHandler defines the actions that a component that does request should do
type DatabaseClientHandler interface {
	DoRequest(req *esapi.IndexRequest) error
	DoBulkRequest(buff *bytes.Buffer, index string) error
	DoBulkRemove(index string, hashes []string) error
	DoMultiGet(hashes []string, index string) (objectsMap, error)

	CheckAndCreateIndex(index string) error
	CheckAndCreateAlias(alias string, index string) error
	CheckAndCreateTemplate(templateName string, template *bytes.Buffer) error
	CheckAndCreatePolicy(policyName string, policy *bytes.Buffer) error

	IsInterfaceNil() bool
}

// DBAccountHandler -
type DBAccountHandler interface {
	GetAccounts(alteredAccounts map[string]*types.AlteredAccount) ([]*types.AccountEGLD, []*types.AccountESDT)
	PrepareAccountsMapEGLD(accounts []*types.AccountEGLD) map[string]*types.AccountInfo
	PrepareAccountsMapESDT(accounts []*types.AccountESDT) map[string]*types.AccountInfo
	PrepareAccountsHistory(accounts map[string]*types.AccountInfo) map[string]*types.AccountBalanceHistory

	SerializeAccountsHistory(accounts map[string]*types.AccountBalanceHistory, bulkSizeThreshold int) ([]bytes.Buffer, error)
	SerializeAccounts(accounts map[string]*types.AccountInfo, bulkSizeThreshold int, areESDTAccounts bool) ([]bytes.Buffer, error)
}

// DBBlockHandler -
type DBBlockHandler interface {
	PrepareBlockForDB(header data.HeaderHandler, signersIndexes []uint64, body *block.Body, notarizedHeadersHashes []string, sizeTxs int) (*types.Block, error)

	SerializeBlock(elasticBlock *types.Block) (*bytes.Buffer, error)
}

// DBTransactionsHandler -
type DBTransactionsHandler interface {
	PrepareTransactionsForDatabase(
		body *block.Body,
		header data.HeaderHandler,
		txPool map[string]data.TransactionHandler,
	) ([]*types.Transaction, []*types.ScResult, []*types.Receipt, map[string]*types.AlteredAccount)

	SetTxLogsProcessor(txLogProcessor process.TransactionLogProcessorDatabase)

	SerializeReceipts(receipts []*types.Receipt, bulkSizeThreshold int) ([]bytes.Buffer, error)
	SerializeTransactions(transactions []*types.Transaction, selfShardID uint32, mbsHashInDB map[string]bool, bulkSizeThreshold int) ([]bytes.Buffer, error)
	SerializeScResults(scResults []*types.ScResult, bulkSizeThreshold int) ([]bytes.Buffer, error)
}

// DBMiniblocksHandler -
type DBMiniblocksHandler interface {
	PrepareDBMiniblocks(header data.HeaderHandler, body *block.Body) []*types.Miniblock

	SerializeBulkMiniBlocks(
		hdrShardID uint32,
		bulkMbs []*types.Miniblock,
		getAlreadyIndexedItems func(hashes []string, index string) (map[string]bool, error),
		miniblocksIndex string,
	) (bytes.Buffer, map[string]bool)
}
