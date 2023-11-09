package postprocess_test

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/postprocess"
	"github.com/stretchr/testify/require"
)

func TestNewFeeAccumulator(t *testing.T) {
	t.Parallel()

	feeHandler, err := postprocess.NewFeeAccumulator()
	require.Nil(t, err)
	require.NotNil(t, feeHandler)
}

func TestFeeHandler_CreateBlockStarted(t *testing.T) {
	t.Parallel()

	feeHandler, _ := postprocess.NewFeeAccumulator()
	feeHandler.ProcessTransactionFee(big.NewInt(100), big.NewInt(50), []byte("txhash"))

	zeroGasAndFees := process.GetZeroGasAndFees()
	feeHandler.CreateBlockStarted(zeroGasAndFees)

	devFees := feeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(0), devFees)

	accumulatedFees := feeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(0), accumulatedFees)
}

func TestFeeHandler_GetAccumulatedFees(t *testing.T) {
	t.Parallel()

	feeHandler, _ := postprocess.NewFeeAccumulator()
	feeHandler.ProcessTransactionFee(big.NewInt(100), big.NewInt(50), []byte("txhash"))

	accumulatedFees := feeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(100), accumulatedFees)
}

func TestFeeHandler_GetDeveloperFees(t *testing.T) {
	t.Parallel()

	feeHandler, _ := postprocess.NewFeeAccumulator()
	feeHandler.ProcessTransactionFee(big.NewInt(100), big.NewInt(50), []byte("txhash"))

	devFees := feeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(50), devFees)
}

func TestFeeHandler_ProcessTransactionFee(t *testing.T) {
	t.Parallel()

	feeHandler, _ := postprocess.NewFeeAccumulator()

	feeHandler.ProcessTransactionFee(big.NewInt(1000), big.NewInt(100), []byte("txhash1"))
	feeHandler.ProcessTransactionFee(big.NewInt(100), big.NewInt(10), []byte("txhash2"))
	feeHandler.ProcessTransactionFee(big.NewInt(10), big.NewInt(1), []byte("txhash3"))

	accFees := feeHandler.GetAccumulatedFees()
	devFees := feeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(1110), accFees)
	require.Equal(t, big.NewInt(111), devFees)
}

func TestFeeHandler_RevertFees(t *testing.T) {
	t.Parallel()

	feeHandler, _ := postprocess.NewFeeAccumulator()

	feeHandler.ProcessTransactionFee(big.NewInt(1000), big.NewInt(100), []byte("txhash1"))
	feeHandler.ProcessTransactionFee(big.NewInt(100), big.NewInt(10), []byte("txhash2"))
	feeHandler.ProcessTransactionFee(big.NewInt(10), big.NewInt(1), []byte("txhash3"))

	feeHandler.RevertFees([][]byte{[]byte("txhash2")})

	accFees := feeHandler.GetAccumulatedFees()
	devFees := feeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(1010), accFees)
	require.Equal(t, big.NewInt(101), devFees)
}

func TestFeeHandler_CompleteRevertFeesUserTxs(t *testing.T) {
	t.Parallel()

	feeHandler, _ := postprocess.NewFeeAccumulator()
	userTxHashes := [][]byte{[]byte("txHash1"), []byte("txHash2"), []byte("txHash3")}
	originalTxHashes := [][]byte{[]byte("origTxHash1"), []byte("origTxHash2"), []byte("origTxHash3")}

	feeHandler.ProcessTransactionFeeRelayedUserTx(big.NewInt(1000), big.NewInt(100), userTxHashes[0], originalTxHashes[0])
	feeHandler.ProcessTransactionFeeRelayedUserTx(big.NewInt(100), big.NewInt(10), userTxHashes[1], originalTxHashes[1])
	feeHandler.ProcessTransactionFeeRelayedUserTx(big.NewInt(10), big.NewInt(1), userTxHashes[2], originalTxHashes[2])

	feeHandler.RevertFees(originalTxHashes)

	accFees := feeHandler.GetAccumulatedFees()
	devFees := feeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(0), accFees)
	require.Equal(t, big.NewInt(0), devFees)
}

func TestFeeHandler_PartialRevertFeesUserTxs(t *testing.T) {
	t.Parallel()
	userTxHashes := [][]byte{[]byte("txHash1"), []byte("txHash2"), []byte("txHash3")}
	originalTxHashes := [][]byte{[]byte("origTxHash1"), []byte("origTxHash2"), []byte("origTxHash3"), []byte("userTxHash4")}

	t.Run("revert partial originalTxs", func(t *testing.T) {
		feeHandler, _ := postprocess.NewFeeAccumulator()
		feeHandler.ProcessTransactionFeeRelayedUserTx(big.NewInt(1000), big.NewInt(100), userTxHashes[0], originalTxHashes[0])
		feeHandler.ProcessTransactionFeeRelayedUserTx(big.NewInt(100), big.NewInt(10), userTxHashes[1], originalTxHashes[1])
		feeHandler.ProcessTransactionFeeRelayedUserTx(big.NewInt(10), big.NewInt(1), userTxHashes[2], originalTxHashes[2])
		feeHandler.ProcessTransactionFee(big.NewInt(2000), big.NewInt(200), originalTxHashes[3])

		feeHandler.RevertFees(originalTxHashes[:3])

		accFees := feeHandler.GetAccumulatedFees()
		devFees := feeHandler.GetDeveloperFees()
		require.Equal(t, big.NewInt(2000), accFees)
		require.Equal(t, big.NewInt(200), devFees)
	})
	t.Run("revert all userTxs", func(t *testing.T) {
		feeHandler, _ := postprocess.NewFeeAccumulator()
		feeHandler.ProcessTransactionFeeRelayedUserTx(big.NewInt(1000), big.NewInt(100), userTxHashes[0], originalTxHashes[0])
		feeHandler.ProcessTransactionFeeRelayedUserTx(big.NewInt(100), big.NewInt(10), userTxHashes[1], originalTxHashes[1])
		feeHandler.ProcessTransactionFeeRelayedUserTx(big.NewInt(10), big.NewInt(1), userTxHashes[2], originalTxHashes[2])
		feeHandler.ProcessTransactionFee(big.NewInt(2000), big.NewInt(200), originalTxHashes[3])

		feeHandler.RevertFees(userTxHashes)

		accFees := feeHandler.GetAccumulatedFees()
		devFees := feeHandler.GetDeveloperFees()
		require.Equal(t, big.NewInt(2000), accFees)
		require.Equal(t, big.NewInt(200), devFees)
	})
	t.Run("revert partial userTxs", func(t *testing.T) {
		feeHandler, _ := postprocess.NewFeeAccumulator()
		feeHandler.ProcessTransactionFeeRelayedUserTx(big.NewInt(1000), big.NewInt(100), userTxHashes[0], originalTxHashes[0])
		feeHandler.ProcessTransactionFeeRelayedUserTx(big.NewInt(100), big.NewInt(10), userTxHashes[1], originalTxHashes[1])
		feeHandler.ProcessTransactionFeeRelayedUserTx(big.NewInt(10), big.NewInt(1), userTxHashes[2], originalTxHashes[2])
		feeHandler.ProcessTransactionFee(big.NewInt(2000), big.NewInt(200), originalTxHashes[3])

		feeHandler.RevertFees(userTxHashes[:2])

		accFees := feeHandler.GetAccumulatedFees()
		devFees := feeHandler.GetDeveloperFees()
		require.Equal(t, big.NewInt(2010), accFees)
		require.Equal(t, big.NewInt(201), devFees)
	})
}

func TestFeeHandler_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	fee, _ := postprocess.NewFeeAccumulator()
	require.False(t, check.IfNil(fee))
}
