package transactionAPI

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
)

type feeComputer interface {
	ComputeTransactionFee(tx data.TransactionWithFeeHandler, epoch int) *big.Int
	ComputeTransactionFeeForMoveBalance(tx data.TransactionWithFeeHandler, epoch int) *big.Int
	IsInterfaceNil() bool
}

// LogsFacade defines the interface of a logs facade
type LogsFacade interface {
	GetLog(logKey []byte, epoch uint32) (*transaction.ApiLogs, error)
	IsInterfaceNil() bool
}
