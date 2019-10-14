package transaction

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
)

// InterceptedTransactionWrapper is a struct which will wrap an InterceptedTransaction in order to implement the
// TxValidatorHandler without needing all the complexity of the other structure
type InterceptedTransactionWrapper struct {
	transaction   *transaction.Transaction
	senderShardId uint32
	feeHandler    process.FeeHandler
}

// NewInterceptedTransactionWrapper will return a new instance of a InterceptedTransactionWrapper
func NewInterceptedTransactionWrapper(
	transaction *transaction.Transaction,
	senderShardId uint32,
	feeHandler process.FeeHandler,
) (*InterceptedTransactionWrapper, error) {
	if transaction == nil {
		return nil, process.ErrNilTransaction
	}
	if feeHandler == nil || feeHandler.IsInterfaceNil() {
		return nil, process.ErrNilEconomicsFeeHandler
	}

	itw := &InterceptedTransactionWrapper{
		transaction:   transaction,
		senderShardId: senderShardId,
		feeHandler:    feeHandler,
	}

	err := itw.checkFees()
	if err != nil {
		return nil, err
	}

	return itw, nil
}

// SenderShardId will return the shard if of the sender
func (itw *InterceptedTransactionWrapper) SenderShardId() uint32 {
	return itw.senderShardId
}

// Nonce will return the nonce specified in the transaction
func (itw *InterceptedTransactionWrapper) Nonce() uint64 {
	return itw.transaction.Nonce
}

// SenderAddress will return the transaction's sender address
func (itw *InterceptedTransactionWrapper) SenderAddress() state.AddressContainer {
	return state.NewAddress(itw.transaction.SndAddr)
}

// TotalValue will return the total value of the transaction, representing the addition of the transaction's value
// and the product of gas price and gas limit
func (itw *InterceptedTransactionWrapper) TotalValue() *big.Int {
	result := big.NewInt(0).Set(itw.transaction.Value)
	gasPrice := big.NewInt(int64(itw.transaction.GasPrice))
	gasLimit := big.NewInt(int64(itw.transaction.GasLimit))
	mulTxCost := big.NewInt(0)
	mulTxCost = mulTxCost.Mul(gasPrice, gasLimit)
	result = result.Add(result, mulTxCost)

	return result
}

func (itw *InterceptedTransactionWrapper) checkFees() error {
	isLowerGasLimitInTx := itw.transaction.GasLimit < itw.feeHandler.MinGasLimit()
	if isLowerGasLimitInTx {
		return process.ErrInsufficientGasLimitInTx
	}

	isLowerGasPrice := itw.transaction.GasPrice < itw.feeHandler.MinGasPrice()
	if isLowerGasPrice {
		return process.ErrInsufficientGasPriceInTx
	}

	return nil
}
