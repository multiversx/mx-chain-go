package mock

// EpochStartTriggerStub -
type EpochStartTriggerStub struct {
	MetaEpochCalled func() uint32
}

// MetaEpoch -
func (e *EpochStartTriggerStub) MetaEpoch() uint32 {
	if e.MetaEpochCalled != nil {
		return e.MetaEpochCalled()
	}
	return 0
}

// IsInterfaceNil -
func (e *EpochStartTriggerStub) IsInterfaceNil() bool {
	return e == nil
}
