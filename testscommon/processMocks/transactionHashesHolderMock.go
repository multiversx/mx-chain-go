package processMocks

// TransactionHashesHolderMock -
type TransactionHashesHolderMock struct {
	AppendCalled       func(hash []byte)
	GetAllHashesCalled func() [][]byte
	ResetCalled        func()
}

// Append -
func (mock *TransactionHashesHolderMock) Append(hash []byte) {
	if mock.AppendCalled != nil {
		mock.AppendCalled(hash)
	}
}

// GetAllHashes -
func (mock *TransactionHashesHolderMock) GetAllHashes() [][]byte {
	if mock.GetAllHashesCalled != nil {
		return mock.GetAllHashesCalled()
	}
	return nil
}

// Reset -
func (mock *TransactionHashesHolderMock) Reset() {
	if mock.ResetCalled != nil {
		mock.ResetCalled()
	}
}

// IsInterfaceNil -
func (mock *TransactionHashesHolderMock) IsInterfaceNil() bool {
	return mock == nil
}
