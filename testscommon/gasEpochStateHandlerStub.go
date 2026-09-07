package testscommon

// GasEpochStateHandlerStub -
type GasEpochStateHandlerStub struct {
	EpochConfirmedCalled                           func(epoch uint32)
	RoundConfirmedCalled                           func(round uint64)
	GetEpochForLimitsAndOverEstimationFactorCalled func() (uint32, uint64)
}

// EpochConfirmed -
func (g *GasEpochStateHandlerStub) EpochConfirmed(epoch uint32) {
	if g.EpochConfirmedCalled != nil {
		g.EpochConfirmedCalled(epoch)
	}
}

// RoundConfirmed -
func (g *GasEpochStateHandlerStub) RoundConfirmed(round uint64) {
	if g.RoundConfirmedCalled != nil {
		g.RoundConfirmedCalled(round)
	}
}

// GetEpochForLimitsAndOverEstimationFactor -
func (g *GasEpochStateHandlerStub) GetEpochForLimitsAndOverEstimationFactor() (uint32, uint64) {
	if g.GetEpochForLimitsAndOverEstimationFactorCalled != nil {
		return g.GetEpochForLimitsAndOverEstimationFactorCalled()
	}

	return 0, 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (g *GasEpochStateHandlerStub) IsInterfaceNil() bool {
	return g == nil
}
