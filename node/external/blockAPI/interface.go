package blockAPI

import (
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
)

// TransactionUnmarshaller defines what a transaction unmarshaler should do
type TransactionUnmarshaller interface {
	UnmarshalTransaction(txBytes []byte, txType transaction.TxType) (*transaction.ApiTransactionResult, error)
	UnmarshalReceipt(receiptBytes []byte) (*transaction.ApiReceipt, error)
	IsInterfaceNil() bool
}

// APIBlockHandler defines the behavior of a component able to return api blocks
type APIBlockHandler interface {
	GetBlockByNonce(nonce uint64, withTxs bool) (*api.Block, error)
	GetBlockByHash(hash []byte, withTxs bool) (*api.Block, error)
	GetBlockByRound(round uint64, withTxs bool) (*api.Block, error)
	IsInterfaceNil() bool
}
