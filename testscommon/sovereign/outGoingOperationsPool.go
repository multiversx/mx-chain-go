package sovereign

// OutGoingOperationsPoolMock -
type OutGoingOperationsPoolMock struct {
	AddCalled func(hash []byte, data []byte)
}

// Add -
func (mock *OutGoingOperationsPoolMock) Add(hash []byte, data []byte) {
	if mock.AddCalled != nil {
		mock.AddCalled(hash, data)
	}
}

// Get -
func (mock *OutGoingOperationsPoolMock) Get(_ []byte) []byte {
	return nil
}

// Delete -
func (mock *OutGoingOperationsPoolMock) Delete(_ []byte) {
}

// GetUnconfirmedOperations -
func (mock *OutGoingOperationsPoolMock) GetUnconfirmedOperations() [][]byte {
	return nil
}

// IsInterfaceNil -
func (mock *OutGoingOperationsPoolMock) IsInterfaceNil() bool {
	return mock == nil
}
