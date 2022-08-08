package txcache

type feeHelper interface {
	gasLimitShift() uint64
	gasPriceShift() uint64
	minPricePerUnit() uint64
	normalizedMinFee() uint64
	minGasPriceFactor() uint64
	IsInterfaceNil() bool
}

type feeComputationHelper struct {
	gasShiftingFactor   uint64
	priceShiftingFactor uint64
	minFeeNormalized    uint64
	minPPUNormalized    uint64
	minPriceFactor      uint64
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
	return fch.minPPUNormalized
}

func (fch *feeComputationHelper) minGasPriceFactor() uint64 {
	return fch.minPriceFactor
}

func (fch *feeComputationHelper) initializeHelperParameters(minPrice, minGasLimit, minPriceProcessing uint64) {
	fch.priceShiftingFactor = computeShiftMagnitude(minPrice, priceBinaryResolution)
	x := minPriceProcessing >> fch.priceShiftingFactor
	for x == 0 && fch.priceShiftingFactor > 0 {
		fch.priceShiftingFactor--
		x = minPriceProcessing >> fch.priceShiftingFactor
	}

	fch.gasShiftingFactor = computeShiftMagnitude(minGasLimit, gasBinaryResolution)

	fch.minPPUNormalized = minPriceProcessing >> fch.priceShiftingFactor
	fch.minFeeNormalized = (minGasLimit >> fch.gasLimitShift()) * (minPrice >> fch.priceShiftingFactor)
	fch.minPriceFactor = minPrice / minPriceProcessing
}

// returns the maximum shift magnitude of the number in order to maintain the given binary resolution
func computeShiftMagnitude(x uint64, resolution uint8) uint64 {
	m := uint64(0)
	stopCondition := uint64(1) << resolution
	shiftStep := uint64(1)

	for i := x; i > stopCondition; i >>= shiftStep {
		m += shiftStep
	}

	return m
}

// IsInterfaceNil returns nil if the underlying object is nil
func (fch *feeComputationHelper) IsInterfaceNil() bool {
	return fch == nil
}
