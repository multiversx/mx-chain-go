package fee

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/data"
)

type economicsDataWithComputeFee interface {
	ComputeTxFee(tx data.TransactionWithFeeHandler) *big.Int
	ComputeMoveBalanceFee(tx data.TransactionWithFeeHandler) *big.Int
}
