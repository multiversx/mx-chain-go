package fee

// testFeeComputer is an exported struct that should be used only in tests
type testFeeComputer struct {
	*feeComputer
}

// NewTestFeeComputer creates a new instance of type testFeeComputer
func NewTestFeeComputer(feeComputerInstance *feeComputer) *testFeeComputer {
	return &testFeeComputer{
		feeComputer: feeComputerInstance,
	}
}

// LenEconomicsInstances returns the number of economic instances
func (computer *testFeeComputer) LenEconomicsInstances() int {
	computer.mutex.RLock()
	defer computer.mutex.RUnlock()

	return len(computer.economicsInstances)
}

// IsInterfaceNil returns true if there is no value under the interface
func (computer *testFeeComputer) IsInterfaceNil() bool {
	return computer == nil
}
