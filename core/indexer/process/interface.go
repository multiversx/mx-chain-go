package process

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
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

// DBAccountHandler defines the actions that an accounts handler should do
type DBAccountHandler interface {
	GetAccounts(alteredAccounts map[string]*types.AlteredAccount) ([]*types.AccountEGLD, []*types.AccountESDT)
	PrepareAccountsMapEGLD(accounts []*types.AccountEGLD) map[string]*types.AccountInfo
	PrepareAccountsMapESDT(accounts []*types.AccountESDT) map[string]*types.AccountInfo
	PrepareAccountsHistory(accounts map[string]*types.AccountInfo) map[string]*types.AccountBalanceHistory

	SerializeAccountsHistory(accounts map[string]*types.AccountBalanceHistory) ([]*bytes.Buffer, error)
	SerializeAccounts(accounts map[string]*types.AccountInfo, areESDTAccounts bool) ([]*bytes.Buffer, error)
}

// DBBlockHandler defines the actions that a block handler should do
type DBBlockHandler interface {
	PrepareBlockForDB(header data.HeaderHandler, signersIndexes []uint64, body *block.Body, notarizedHeadersHashes []string, sizeTxs int) (*types.Block, error)
	ComputeHeaderHash(header data.HeaderHandler) ([]byte, error)

	SerializeEpochInfoData(header data.HeaderHandler) (*bytes.Buffer, error)
	SerializeBlock(elasticBlock *types.Block) (*bytes.Buffer, error)
}

// DBTransactionsHandler defines the actions that a transactions handler should do
type DBTransactionsHandler interface {
	PrepareTransactionsForDatabase(
		body *block.Body,
		header data.HeaderHandler,
		txPool map[string]data.TransactionHandler,
	) *types.PreparedResults
	GetRewardsTxsHashesHexEncoded(header data.HeaderHandler, body *block.Body) []string

	SetTxLogsProcessor(txLogProcessor process.TransactionLogProcessorDatabase)

	SerializeReceipts(receipts []*types.Receipt) ([]*bytes.Buffer, error)
	SerializeTransactions(transactions []*types.Transaction, selfShardID uint32, mbsHashInDB map[string]bool) ([]*bytes.Buffer, error)
	SerializeScResults(scResults []*types.ScResult) ([]*bytes.Buffer, error)
}

// DBMiniblocksHandler defines the actions that a miniblocks handler should do
type DBMiniblocksHandler interface {
	PrepareDBMiniblocks(header data.HeaderHandler, body *block.Body) []*types.Miniblock
	GetMiniblocksHashesHexEncoded(header data.HeaderHandler, body *block.Body) []string

	SerializeBulkMiniBlocks(bulkMbs []*types.Miniblock, mbsInDB map[string]bool) *bytes.Buffer
}

// DBGeneralInfoHandler defines the actions that a general info handler should do
type DBGeneralInfoHandler interface {
	PrepareGeneralInfo(tpsBenchmark statistics.TPSBenchmark) (*types.TPS, []*types.TPS)

	SerializeGeneralInfo(genInfo *types.TPS, shardsInfo []*types.TPS, index string) *bytes.Buffer
	SerializeRoundsInfo(roundsInfo []types.RoundInfo) *bytes.Buffer
}

// DBValidatorsHandler defines the actions that a validators handler should do
type DBValidatorsHandler interface {
	PrepareValidatorsPublicKeys(shardValidatorsPubKeys [][]byte) *types.ValidatorsPublicKeys
	SerializeValidatorsPubKeys(validatorsPubKeys *types.ValidatorsPublicKeys) (*bytes.Buffer, error)
	SerializeValidatorsRating(index string, validatorsRatingInfo []types.ValidatorRatingInfo) ([]*bytes.Buffer, error)
}
