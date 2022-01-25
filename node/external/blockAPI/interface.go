package blockAPI

import (
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
)

// TransactionUnmarshalerHandler defines the transaction unmarshaler handler should do
type TransactionUnmarshalerHandler interface {
	UnmarshalTransaction(txBytes []byte, txType transaction.TxType) (*transaction.ApiTransactionResult, error)
}
