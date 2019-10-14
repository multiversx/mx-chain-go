package transaction_test

import (
	"math/big"
	"testing"

	erdTx "github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/stretchr/testify/assert"
)

func TestNewInterceptedTransactionWrapper_NilTxShouldErr(t *testing.T) {
	t.Parallel()

	itw, err := transaction.NewInterceptedTransactionWrapper(nil, 0, &mock.FeeHandlerStub{})
	assert.Nil(t, itw)
	assert.Equal(t, process.ErrNilTransaction, err)
}

func TestNewInterceptedTransactionWrapper_NilFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	tx := erdTx.Transaction{Nonce: 1}
	itw, err := transaction.NewInterceptedTransactionWrapper(&tx, 0, nil)
	assert.Nil(t, itw)
	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestNewInterceptedTransactionWrapper_InsufficientGasLimitShouldErr(t *testing.T) {
	t.Parallel()

	tx := erdTx.Transaction{
		Nonce:    1,
		GasPrice: 0,
		GasLimit: 2,
	}
	itw, err := transaction.NewInterceptedTransactionWrapper(&tx, 0, getFeeHandler())
	assert.Nil(t, itw)
	assert.Equal(t, process.ErrInsufficientGasLimitInTx, err)
}

func TestNewInterceptedTransactionWrapper_ShouldWork(t *testing.T) {
	t.Parallel()

	tx := erdTx.Transaction{
		Nonce:    1,
		GasPrice: 0,
		GasLimit: 5,
	}
	itw, err := transaction.NewInterceptedTransactionWrapper(&tx, 0, getFeeHandler())
	assert.NotNil(t, itw)
	assert.Nil(t, err)
}

func TestInterceptedTransactionWrapper_Nonce(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	tx := erdTx.Transaction{
		Nonce:    nonce,
		GasPrice: 0,
		GasLimit: 5,
	}
	itw, err := transaction.NewInterceptedTransactionWrapper(&tx, 0, getFeeHandler())
	assert.Nil(t, err)
	assert.NotNil(t, itw)

	assert.Equal(t, nonce, itw.Nonce())
}

func TestInterceptedTransactionWrapper_SenderShardId(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	senderShardId := uint32(3)
	tx := erdTx.Transaction{
		Nonce:    nonce,
		GasPrice: 0,
		GasLimit: 5,
	}
	itw, _ := transaction.NewInterceptedTransactionWrapper(&tx, senderShardId, getFeeHandler())
	assert.NotNil(t, itw)

	assert.Equal(t, senderShardId, itw.SenderShardId())
}

func TestInterceptedTransactionWrapper_TotalValue(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	senderShardId := uint32(3)
	gasPrice := uint64(5)
	gasLimit := uint64(100)
	txValue := big.NewInt(2000)

	// 2000 + 5 * 100
	expectedTxValue := big.NewInt(2500)
	tx := erdTx.Transaction{
		Nonce:    nonce,
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		Value:    txValue,
	}
	itw, _ := transaction.NewInterceptedTransactionWrapper(&tx, senderShardId, getFeeHandler())
	assert.NotNil(t, itw)

	assert.Equal(t, expectedTxValue, itw.TotalValue())
}

func getFeeHandler() process.FeeHandler {
	feeHandler := mock.FeeHandlerStub{
		MinGasLimitCalled: func() uint64 {
			return uint64(5)
		},
		MinGasPriceCalled: func() uint64 {
			return uint64(0)
		},
	}

	return &feeHandler
}
