package sovereign

import "github.com/multiversx/mx-chain-core-go/data/sovereign"

// OutGoingOperationsPoolMock -
type OutGoingOperationsPoolMock struct {
	AddCalled                      func(data *sovereign.BridgeOutGoingData)
	GetCalled                      func(hash []byte) *sovereign.BridgeOutGoingData
	DeleteCalled                   func(hash []byte)
	GetUnconfirmedOperationsCalled func() []*sovereign.BridgeOutGoingData
	ConfirmOperationCalled         func(hashOfHashes []byte, hash []byte) error
}

// Add -
func (mock *OutGoingOperationsPoolMock) Add(data *sovereign.BridgeOutGoingData) {
	if mock.AddCalled != nil {
		mock.AddCalled(data)
	}
}

// Get -
func (mock *OutGoingOperationsPoolMock) Get(hash []byte) *sovereign.BridgeOutGoingData {
	if mock.GetCalled != nil {
		return mock.GetCalled(hash)
	}
	return nil
}

// Delete -
func (mock *OutGoingOperationsPoolMock) Delete(hash []byte) {
	if mock.DeleteCalled != nil {
		mock.DeleteCalled(hash)
	}
}

// GetUnconfirmedOperations -
func (mock *OutGoingOperationsPoolMock) GetUnconfirmedOperations() []*sovereign.BridgeOutGoingData {
	if mock.GetUnconfirmedOperationsCalled != nil {
		return mock.GetUnconfirmedOperationsCalled()
	}
	return nil
}

// ConfirmOperation -
func (mock *OutGoingOperationsPoolMock) ConfirmOperation(hashOfHashes []byte, hash []byte) error {
	if mock.ConfirmOperationCalled != nil {
		return mock.ConfirmOperationCalled(hashOfHashes, hash)
	}
	return nil
}

// IsInterfaceNil -
func (mock *OutGoingOperationsPoolMock) IsInterfaceNil() bool {
	return mock == nil
}
