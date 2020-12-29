package txcache

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func Test_estimateTxFeeScore(t *testing.T) {
	txGasHandler, txFeeHelper := dummyParamsWithGasPrice(100 * oneBillion)
	A := createTxWithParams([]byte("a"), "a", 1, 200, 50000, 100*oneBillion)
	B := createTxWithParams([]byte("b"), "b", 1, 200, 50000000, 100*oneBillion)
	C := createTxWithParams([]byte("C"), "c", 1, 200, 1500000000, 100*oneBillion)

	scoreA := estimateTxFeeScore(A, txGasHandler, txFeeHelper)
	scoreB := estimateTxFeeScore(B, txGasHandler, txFeeHelper)
	scoreC := estimateTxFeeScore(C, txGasHandler, txFeeHelper)
	require.Equal(t, uint64(8940), scoreA)
	require.Equal(t, uint64(8940), A.TxFeeScoreNormalized)
	require.Equal(t, uint64(6837580), scoreB)
	require.Equal(t, uint64(6837580), B.TxFeeScoreNormalized)
	require.Equal(t, uint64(205079820), scoreC)
	require.Equal(t, uint64(205079820), C.TxFeeScoreNormalized)
}

func Test_normalizeGasPriceProcessing(t *testing.T) {
	txGasHandler, txFeeHelper := dummyParamsWithGasPriceAndDivisor(100*oneBillion, 100)
	A := createTxWithParams([]byte("A"), "a", 1, 200, 1500000000, 100*oneBillion)
	normalizedGasPriceProcess := normalizeGasPriceProcessing(A, txGasHandler, txFeeHelper)
	require.Equal(t, uint64(7), normalizedGasPriceProcess)

	txGasHandler, txFeeHelper = dummyParamsWithGasPriceAndDivisor(100*oneBillion, 50)
	normalizedGasPriceProcess = normalizeGasPriceProcessing(A, txGasHandler, txFeeHelper)
	require.Equal(t, uint64(14), normalizedGasPriceProcess)

	txGasHandler, txFeeHelper = dummyParamsWithGasPriceAndDivisor(100*oneBillion, 1)
	normalizedGasPriceProcess = normalizeGasPriceProcessing(A, txGasHandler, txFeeHelper)
	require.Equal(t, uint64(745), normalizedGasPriceProcess)

	txGasHandler, txFeeHelper = dummyParamsWithGasPriceAndDivisor(100000, 100)
	A = createTxWithParams([]byte("A"), "a", 1, 200, 1500000000, 100000)
	normalizedGasPriceProcess = normalizeGasPriceProcessing(A, txGasHandler, txFeeHelper)
	require.Equal(t, uint64(7), normalizedGasPriceProcess)
}

func Test_computeProcessingGasPriceAdjustment(t *testing.T) {
	txGasHandler, txFeeHelper := dummyParamsWithGasPriceAndDivisor(100*oneBillion, 100)
	A := createTxWithParams([]byte("A"), "a", 1, 200, 1500000000, 100*oneBillion)
	adjustment := computeProcessingGasPriceAdjustment(A, txGasHandler, txFeeHelper)
	require.Equal(t, uint64(80), adjustment)

	A = createTxWithParams([]byte("A"), "a", 1, 200, 1500000000, 150*oneBillion)
	adjustment = computeProcessingGasPriceAdjustment(A, txGasHandler, txFeeHelper)
	expectedAdjustment := float64(100) * processFeeFactor / float64(1.5)
	require.Equal(t, uint64(expectedAdjustment), adjustment)

	A = createTxWithParams([]byte("A"), "a", 1, 200, 1500000000, 110*oneBillion)
	adjustment = computeProcessingGasPriceAdjustment(A, txGasHandler, txFeeHelper)
	expectedAdjustment = float64(100) * processFeeFactor / float64(1.1)
	require.Equal(t, uint64(expectedAdjustment), adjustment)
}

func dummyParamsWithGasPriceAndDivisor(minGasPrice, processingPriceDivisor uint64) (TxGasHandler, feeHelper) {
	minPrice := minGasPrice
	minPriceProcessing := minGasPrice / processingPriceDivisor
	minGasLimit := uint64(50000)
	txFeeHelper := newFeeComputationHelper(minPrice, minGasLimit, minPriceProcessing)
	txGasHandler := &txcachemocks.TxGasHandlerMock{
		MinimumGasMove:       minGasLimit,
		MinimumGasPrice:      minPrice,
		GasProcessingDivisor: processingPriceDivisor,
	}
	return txGasHandler, txFeeHelper
}
