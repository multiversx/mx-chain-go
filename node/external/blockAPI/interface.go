package blockAPI

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/common"
)

// APITransactionHandler defines what a transaction handler should do
type APITransactionHandler interface {
	UnmarshalTransaction(txBytes []byte, txType transaction.TxType) (*transaction.ApiTransactionResult, error)
	UnmarshalReceipt(receiptBytes []byte) (*transaction.ApiReceipt, error)
	PopulateComputedFields(tx *transaction.ApiTransactionResult)
	IsInterfaceNil() bool
}

// APIBlockHandler defines the behavior of a component able to return api blocks
type APIBlockHandler interface {
	GetBlockByNonce(nonce uint64, options api.BlockQueryOptions) (*api.Block, error)
	GetBlockByHash(hash []byte, options api.BlockQueryOptions) (*api.Block, error)
	GetBlockByRound(round uint64, options api.BlockQueryOptions) (*api.Block, error)
	GetAlteredAccountsForBlock(options api.GetAlteredAccountsForBlockOptions) ([]*outport.AlteredAccount, error)
	IsInterfaceNil() bool
}

// APIInternalBlockHandler defines the behaviour of a component able to return internal blocks
type APIInternalBlockHandler interface {
	GetInternalShardBlockByNonce(format common.ApiOutputFormat, nonce uint64) (interface{}, error)
	GetInternalShardBlockByHash(format common.ApiOutputFormat, hash []byte) (interface{}, error)
	GetInternalShardBlockByRound(format common.ApiOutputFormat, round uint64) (interface{}, error)
	GetInternalMetaBlockByNonce(format common.ApiOutputFormat, nonce uint64) (interface{}, error)
	GetInternalMetaBlockByHash(format common.ApiOutputFormat, hash []byte) (interface{}, error)
	GetInternalMetaBlockByRound(format common.ApiOutputFormat, round uint64) (interface{}, error)
	GetInternalStartOfEpochMetaBlock(format common.ApiOutputFormat, epoch uint32) (interface{}, error)
	GetInternalStartOfEpochValidatorsInfo(format common.ApiOutputFormat, epoch uint32) (interface{}, error)
	GetInternalMiniBlock(format common.ApiOutputFormat, hash []byte, epoch uint32) (interface{}, error)
	IsInterfaceNil() bool
}

type logsFacade interface {
	IncludeLogsInTransactions(txs []*transaction.ApiTransactionResult, logsKeys [][]byte, epoch uint32) error
	IsInterfaceNil() bool
}

type receiptsRepository interface {
	LoadReceipts(header data.HeaderHandler, headerHash []byte) (common.ReceiptsHolder, error)
	IsInterfaceNil() bool
}
