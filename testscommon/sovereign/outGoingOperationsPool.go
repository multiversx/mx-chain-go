package sovereign

// OutGoingOperationsPoolStub -
type OutGoingOperationsPoolStub struct {
}

// Add -
func (stub *OutGoingOperationsPoolStub) Add(_ []byte, _ []byte) {
}

// Get -
func (stub *OutGoingOperationsPoolStub) Get(_ []byte) []byte {
	return nil
}

// Delete -
func (stub *OutGoingOperationsPoolStub) Delete(_ []byte) {
}

// GetUnconfirmedOperations -
func (stub *OutGoingOperationsPoolStub) GetUnconfirmedOperations() [][]byte {
	return nil
}

// IsInterfaceNil -
func (stub *OutGoingOperationsPoolStub) IsInterfaceNil() bool {
	return stub == nil
}
