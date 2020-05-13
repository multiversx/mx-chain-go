package postprocess_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process/block/postprocess"
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
	feeHandler.CreateBlockStarted()

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

func TestFeeHandler_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	fee, _ := postprocess.NewFeeAccumulator()
	require.False(t, check.IfNil(fee))
}
