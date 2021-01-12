package txsFee

import (
	"errors"
	"math"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// minGasPrice = 1, gasPerDataByte = 1, minGasLimit = 1
func TestMoveBalanceSelfShouldWorkAndConsumeTxFee(t *testing.T) {
	testContext := vm.CreatePreparedTxProcessorWithVMs(t, vm.ArgEnableEpoch{})
	defer testContext.Close()

	sndAddr := []byte("12345678901234567890123456789012")
	senderNonce := uint64(0)
	senderBalance := big.NewInt(10000)
	gasPrice := uint64(10)
	gasLimit := uint64(100)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)
	tx := vm.CreateTransaction(senderNonce, big.NewInt(100), sndAddr, sndAddr, gasPrice, gasLimit, []byte("aaaa"))

	_, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Nil(t, testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	// 10_000 - gasPrice(10)*gasLimit(1) + gasPerDataByte(1)*gasPrice(10) = 9950
	expectedBalance := big.NewInt(9950)
	vm.TestAccount(t, testContext.Accounts, sndAddr, senderNonce+1, expectedBalance)

	//check receipts
	require.Equal(t, 1, len(testContext.GetIntermediateTransactions(t)))
	rcpt := testContext.GetIntermediateTransactions(t)[0].(*receipt.Receipt)
	assert.Equal(t, "950", rcpt.Value.String())

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(50), accumulatedFees)
}

// minGasPrice = 1, gasPerDataByte = 1, minGasLimit = 1
func TestMoveBalanceAllFlagsEnabledLessBalanceThanGasLimitMulGasPrice(t *testing.T) {
	testContext := vm.CreatePreparedTxProcessorWithVMs(t, vm.ArgEnableEpoch{})
	defer testContext.Close()

	sndAddr := []byte("12345678901234567890123456789012")
	senderNonce := uint64(0)
	senderBalance := big.NewInt(10000)
	gasPrice := uint64(10)
	gasLimit := uint64(10000)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)
	tx := vm.CreateTransaction(senderNonce, big.NewInt(100), sndAddr, sndAddr, gasPrice, gasLimit, []byte("aaaa"))

	_, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.True(t, errors.Is(err, process.ErrInsufficientFee))
}

func TestMoveBalanceShouldWork(t *testing.T) {
	testContext := vm.CreatePreparedTxProcessorWithVMs(t, vm.ArgEnableEpoch{})
	defer testContext.Close()

	sndAddr := []byte("12345678901234567890123456789012")
	rcvAddr := []byte("12345678901234567890123456789022")
	senderNonce := uint64(0)
	senderBalance := big.NewInt(10000)
	gasPrice := uint64(10)
	gasLimit := uint64(100)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)
	tx := vm.CreateTransaction(senderNonce, big.NewInt(100), sndAddr, rcvAddr, gasPrice, gasLimit, []byte("aaaa"))

	_, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Nil(t, testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	// 10_000 - gasPrice(10)*gasLimit(1) + gasPerDataByte(1)*gasPrice(10) - transferredValue(100) = 9850
	// verify balance of sender
	expectedBalance := big.NewInt(9850)
	vm.TestAccount(
		t,
		testContext.Accounts,
		sndAddr,
		senderNonce+1,
		expectedBalance)

	//verify receiver
	expectedBalanceReceiver := big.NewInt(100)
	vm.TestAccount(t, testContext.Accounts, rcvAddr, 0, expectedBalanceReceiver)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(50), accumulatedFees)
}

func TestMoveBalanceInvalidHasGasButNoValueShouldConsumeGas(t *testing.T) {
	testContext := vm.CreatePreparedTxProcessorWithVMs(t, vm.ArgEnableEpoch{})
	defer testContext.Close()

	sndAddr := []byte("12345678901234567890123456789012")
	rcvAddr := []byte("12345678901234567890123456789022")
	senderBalance := big.NewInt(100)
	gasPrice := uint64(1)
	gasLimit := uint64(20)
	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)
	tx := vm.CreateTransaction(0, big.NewInt(100), sndAddr, rcvAddr, gasPrice, gasLimit, []byte("aaaa"))

	_, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, process.ErrFailedTransaction, err)
	require.Nil(t, testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	// check sender balance
	expectedBalance := big.NewInt(80)
	vm.TestAccount(t, testContext.Accounts, sndAddr, 1, expectedBalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(20), accumulatedFees)
}

func TestMoveBalanceHigherNonceShouldNotConsumeGas(t *testing.T) {
	testContext := vm.CreatePreparedTxProcessorWithVMs(t, vm.ArgEnableEpoch{})
	defer testContext.Close()

	sndAddr := []byte("12345678901234567890123456789012")
	rcvAddr := []byte("12345678901234567890123456789022")

	senderBalance := big.NewInt(100)
	gasPrice := uint64(1)
	gasLimit := uint64(20)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)
	tx := vm.CreateTransaction(1, big.NewInt(500), sndAddr, rcvAddr, gasPrice, gasLimit, []byte("aaaa"))

	_, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, process.ErrHigherNonceInTransaction, err)
	require.Nil(t, testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	// check sender balance
	expectedBalance := big.NewInt(0).Set(senderBalance)
	vm.TestAccount(t, testContext.Accounts, sndAddr, 0, expectedBalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(0), accumulatedFees)
}

func TestMoveBalanceMoreGasThanGasLimitPerBlock(t *testing.T) {
	testContext := vm.CreatePreparedTxProcessorWithVMs(t, vm.ArgEnableEpoch{})
	defer testContext.Close()

	sndAddr := []byte("12345678901234567890123456789012")
	rcvAddr := []byte("12345678901234567890123456789022")

	senderBalance := big.NewInt(0).SetUint64(math.MaxUint64)
	gasPrice := uint64(1)
	gasLimit := uint64(math.MaxUint64)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)
	tx := vm.CreateTransaction(0, big.NewInt(500), sndAddr, rcvAddr, gasPrice, gasLimit, []byte("aaaa"))

	_, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, process.ErrMoreGasThanGasLimitPerBlock, err)
	require.Nil(t, testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	// check sender balance
	expectedBalance := big.NewInt(0).Set(senderBalance)
	vm.TestAccount(t, testContext.Accounts, sndAddr, 0, expectedBalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(0), accumulatedFees)
}
