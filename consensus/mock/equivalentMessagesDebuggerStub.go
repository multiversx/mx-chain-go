package mock

// EquivalentMessagesDebuggerStub -
type EquivalentMessagesDebuggerStub struct {
	DisplayEquivalentMessagesStatisticsCalled func(getDataHandler func() map[string]uint64)
}

// DisplayEquivalentMessagesStatistics -
func (stub *EquivalentMessagesDebuggerStub) DisplayEquivalentMessagesStatistics(getDataHandler func() map[string]uint64) {
	if stub.DisplayEquivalentMessagesStatisticsCalled != nil {
		stub.DisplayEquivalentMessagesStatisticsCalled(getDataHandler)
	}
}

// IsInterfaceNil -
func (stub *EquivalentMessagesDebuggerStub) IsInterfaceNil() bool {
	return stub == nil
}
