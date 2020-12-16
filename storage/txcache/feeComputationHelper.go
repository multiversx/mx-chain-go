package txcache

type feeHelper interface {
	gasLimitShift() uint64
	gasPriceShift() uint64
	minPricePerUnit() uint64
	normalizedMinFee() uint64
	IsInterfaceNil() bool
}

type feeComputationHelper struct {
	gasShiftingFactor   uint64
	priceShiftingFactor uint64
	minFeeNormalized    uint64
	minPPU              uint64
}

const priceBinaryResolution = 10
const gasBinaryResolution = 4

func newFeeComputationHelper(minPrice, minGasLimit, minPriceProcessing uint64) *feeComputationHelper {
	feeComputeHelper := &feeComputationHelper{}
	feeComputeHelper.initializeHelperParameters(minPrice, minGasLimit, minPriceProcessing)
	return feeComputeHelper
}

func (fch *feeComputationHelper) gasLimitShift() uint64 {
	return fch.gasShiftingFactor
}

func (fch *feeComputationHelper) gasPriceShift() uint64 {
	return fch.priceShiftingFactor
}

func (fch *feeComputationHelper) normalizedMinFee() uint64 {
	return fch.minFeeNormalized
}

func (fch *feeComputationHelper) minPricePerUnit() uint64 {
	return fch.minPPU
}

func (fch *feeComputationHelper) initializeHelperParameters(minPrice, minGasLimit, minPriceProcessing uint64) {
	fch.priceShiftingFactor = binaryMagnitude(minPrice)
	if fch.priceShiftingFactor > priceBinaryResolution {
		fch.priceShiftingFactor = fch.priceShiftingFactor - priceBinaryResolution
	}
	for x := minPriceProcessing >> fch.priceShiftingFactor; x == 0; {
		fch.priceShiftingFactor--
	}

	fch.gasShiftingFactor = binaryMagnitude(minGasLimit)
	if fch.gasShiftingFactor > gasBinaryResolution {
		fch.gasShiftingFactor = fch.gasShiftingFactor - gasBinaryResolution
	}

	fch.minPPU = minPriceProcessing >> fch.priceShiftingFactor
	fch.minFeeNormalized = (minGasLimit >> fch.gasLimitShift()) * (minPrice >> fch.priceShiftingFactor)
}

// returns the binary order of the number
// eg for 1024 it should return 10
func binaryMagnitude(x uint64) uint64 {
	m := uint64(1)

	for i := x; i > 1; i >>= 1 {
		m <<= 1
	}
	return m
}

// IsInterfaceNil returns nil if the underlying object is nil
func (fch *feeComputationHelper) IsInterfaceNil() bool {
	return fch == nil
}
