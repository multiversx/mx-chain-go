package transactionAPI

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
)

type feeComputer interface {
	ComputeTransactionFee(tx data.TransactionWithFeeHandler, epoch int) *big.Int
	IsInterfaceNil() bool
}

// LogsRepository defines the interface of a logs repository
type LogsRepository interface {
	GetLog(txHash []byte, epoch uint32) (*transaction.ApiLogs, error)
	IsInterfaceNil() bool
}
