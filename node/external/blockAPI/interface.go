package blockAPI

import (
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
)

// TransactionUnmarshalerHandler defines the transaction unmarshaler handler should do
type TransactionUnmarshalerHandler interface {
	UnmarshalTransaction(txBytes []byte, txType transaction.TxType) (*transaction.ApiTransactionResult, error)
}

// APIBlockHandler defines the behavior of a component able to return api blocks
type APIBlockHandler interface {
	GetBlockByNonce(nonce uint64, withTxs bool) (*api.Block, error)
	GetBlockByHash(hash []byte, withTxs bool) (*api.Block, error)
	GetBlockByRound(round uint64, withTxs bool) (*api.Block, error)
	IsInterfaceNil() bool
}
