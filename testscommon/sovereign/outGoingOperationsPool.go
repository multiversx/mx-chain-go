package sovereign

// OutGoingOperationsPoolStub -
type OutGoingOperationsPoolStub struct {
	AddCalled func(hash []byte, data []byte)
}

// Add -
func (stub *OutGoingOperationsPoolStub) Add(hash []byte, data []byte) {
	if stub.AddCalled != nil {
		stub.AddCalled(hash, data)
	}
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
