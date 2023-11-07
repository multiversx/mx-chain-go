package common

// TxExecutionOrderHandlerStub -
type TxExecutionOrderHandlerStub struct {
	AddCalled            func(txHash []byte)
	GetItemAtIndexCalled func(index uint32) ([]byte, error)
	GetOrderCalled       func(txHash []byte) (int, error)
	RemoveCalled         func(txHash []byte)
	RemoveMultipleCalled func(txHashes [][]byte)
	GetItemsCalled       func() [][]byte
	ContainsCalled       func(txHash []byte) bool
	ClearCalled          func()
	LenCalled            func() int
}

// Add -
func (teohs *TxExecutionOrderHandlerStub) Add(txHash []byte) {
	if teohs.AddCalled != nil {
		teohs.AddCalled(txHash)
	}
}

// GetItemAtIndex -
func (teohs *TxExecutionOrderHandlerStub) GetItemAtIndex(index uint32) ([]byte, error) {
	if teohs.GetItemAtIndexCalled != nil {
		return teohs.GetItemAtIndexCalled(index)
	}
	return nil, nil
}

// GetOrder -
func (teohs *TxExecutionOrderHandlerStub) GetOrder(txHash []byte) (int, error) {
	if teohs.GetOrderCalled != nil {
		return teohs.GetOrderCalled(txHash)
	}
	return 0, nil
}

// Remove -
func (teohs *TxExecutionOrderHandlerStub) Remove(txHash []byte) {
	if teohs.RemoveCalled != nil {
		teohs.RemoveCalled(txHash)
	}
}

// RemoveMultiple -
func (teohs *TxExecutionOrderHandlerStub) RemoveMultiple(txHashes [][]byte) {
	if teohs.RemoveMultipleCalled != nil {
		teohs.RemoveMultipleCalled(txHashes)
	}
}

// GetItems -
func (teohs *TxExecutionOrderHandlerStub) GetItems() [][]byte {
	if teohs.GetItemsCalled != nil {
		return teohs.GetItemsCalled()
	}
	return nil
}

// Contains -
func (teohs *TxExecutionOrderHandlerStub) Contains(txHash []byte) bool {
	if teohs.ContainsCalled != nil {
		return teohs.ContainsCalled(txHash)
	}
	return false
}

// Clear -
func (teohs *TxExecutionOrderHandlerStub) Clear() {
	if teohs.ClearCalled != nil {
		teohs.ClearCalled()
	}
}

// Len -
func (teohs *TxExecutionOrderHandlerStub) Len() int {
	if teohs.LenCalled != nil {
		return teohs.LenCalled()
	}
	return 0
}

// IsInterfaceNil -
func (teohs *TxExecutionOrderHandlerStub) IsInterfaceNil() bool {
	return teohs == nil
}
