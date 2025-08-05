package epochproviders

import "github.com/multiversx/mx-chain-go/process"

// NewTestArithmeticEpochProvider -
func NewTestArithmeticEpochProvider(arg ArgArithmeticEpochProvider, unixHandler func() int64) *arithmeticEpochProvider {
	aep := &arithmeticEpochProvider{
		headerEpoch:                0,
		headerTimestampForNewEpoch: uint64(arg.StartTime),
		chainParamsHandler:         arg.ChainParametersHandler,
		getUnixHandler:             unixHandler,
		enableEpochsHandler:        arg.EnableEpochsHandler,
	}
	aep.computeCurrentEpoch()

	return aep
}

// CurrentComputedEpoch -
func (aep *arithmeticEpochProvider) CurrentComputedEpoch() uint32 {
	aep.RLock()
	defer aep.RUnlock()

	return aep.currentComputedEpoch
}

// SetUnixHandler -
func (aep *arithmeticEpochProvider) SetUnixHandler(handler func() int64) {
	aep.Lock()
	defer aep.Unlock()

	aep.getUnixHandler = handler
}

// SetCurrentComputedEpoch -
func (aep *arithmeticEpochProvider) SetCurrentComputedEpoch(epoch uint32) {
	aep.Lock()
	defer aep.Unlock()

	aep.currentComputedEpoch = epoch
}

// SetChainParametersHandler -
func (aep *arithmeticEpochProvider) SetChainParametersHandler(chainParamsHandler process.ChainParametersHandler) {
	aep.chainParamsHandler = chainParamsHandler
}
