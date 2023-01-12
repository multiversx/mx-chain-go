package fee

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
)

type economicsDataWithComputeFee interface {
	ComputeTxFee(tx data.TransactionWithFeeHandler) *big.Int
}
