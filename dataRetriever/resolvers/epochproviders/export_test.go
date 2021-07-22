package epochproviders

// NewTestArithmeticEpochProvider -
func NewTestArithmeticEpochProvider(arg ArgArithmeticEpochProvider, unixHandler func() int64) *arithmeticEpochProvider {
	aep := &arithmeticEpochProvider{
		headerEpoch:                0,
		headerTimestampForNewEpoch: uint64(arg.StartTime),
		roundsPerEpoch:             arg.RoundsPerEpoch,
		roundTimeInMilliseconds:    arg.RoundTimeInMilliseconds,
		startTime:                  arg.StartTime,
		getUnixHandler:             unixHandler,
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
