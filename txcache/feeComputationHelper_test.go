package txcache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_initializeHelperParameters(t *testing.T) {
	fch := &feeComputationHelper{
		gasShiftingFactor:   0,
		priceShiftingFactor: 0,
		minFeeNormalized:    0,
		minPPUNormalized:    0,
		minPriceFactor:      0,
	}

	fch.initializeHelperParameters(1<<20, 1<<10, 1<<10)
	require.Equal(t, uint64(10), fch.priceShiftingFactor)
	require.Equal(t, uint64(6), fch.gasShiftingFactor)
	require.Equal(t, uint64(1<<10), fch.minPriceFactor)
	require.Equal(t, uint64((1<<4)*(1<<10)), fch.minFeeNormalized)
	require.Equal(t, uint64(1), fch.minPPUNormalized)

	fch.initializeHelperParameters(1<<22, 1<<17, 1<<7)
	require.Equal(t, uint64(7), fch.priceShiftingFactor)
	require.Equal(t, uint64(13), fch.gasShiftingFactor)
	require.Equal(t, uint64(1<<15), fch.minPriceFactor)
	require.Equal(t, uint64((1<<4)*(1<<15)), fch.minFeeNormalized)
	require.Equal(t, uint64(1), fch.minPPUNormalized)

	fch.initializeHelperParameters(1<<20, 1<<3, 1<<15)
	require.Equal(t, uint64(10), fch.priceShiftingFactor)
	require.Equal(t, uint64(0), fch.gasShiftingFactor)
	require.Equal(t, uint64(1<<5), fch.minPriceFactor)
	require.Equal(t, uint64((1<<3)*(1<<10)), fch.minFeeNormalized)
	require.Equal(t, uint64(1<<5), fch.minPPUNormalized)
}

func Test_newFeeComputationHelper(t *testing.T) {
	fch := newFeeComputationHelper(1<<20, 1<<10, 1<<10)
	require.Equal(t, uint64(10), fch.priceShiftingFactor)
	require.Equal(t, uint64(6), fch.gasShiftingFactor)
	require.Equal(t, uint64(1<<10), fch.minPriceFactor)
	require.Equal(t, uint64((1<<4)*(1<<10)), fch.minFeeNormalized)
	require.Equal(t, uint64(1), fch.minPPUNormalized)
}

func Test_getters(t *testing.T) {
	fch := newFeeComputationHelper(1<<20, 1<<10, 1<<10)
	gasShift := fch.gasLimitShift()
	gasPriceShift := fch.gasPriceShift()
	minFeeNormalized := fch.normalizedMinFee()
	minPPUNormalized := fch.minPricePerUnit()
	minGasPriceFactor := fch.minGasPriceFactor()

	require.Equal(t, uint64(10), gasPriceShift)
	require.Equal(t, uint64(6), gasShift)
	require.Equal(t, uint64(1<<10), minGasPriceFactor)
	require.Equal(t, uint64((1<<4)*(1<<10)), minFeeNormalized)
	require.Equal(t, uint64(1), minPPUNormalized)
}

func Test_computeShiftMagnitude(t *testing.T) {
	shift := computeShiftMagnitude(1<<20, 10)
	require.Equal(t, uint64(10), shift)

	shift = computeShiftMagnitude(1<<12, 10)
	require.Equal(t, uint64(2), shift)

	shift = computeShiftMagnitude(1<<8, 10)
	require.Equal(t, uint64(0), shift)
}
