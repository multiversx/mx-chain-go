package transactionAPI

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/data"
)

type feeComputer interface {
	ComputeTransactionFee(tx data.TransactionWithFeeHandler, epoch int) *big.Int
	IsInterfaceNil() bool
}
