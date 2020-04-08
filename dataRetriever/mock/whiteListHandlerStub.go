package mock

// WhiteListHandlerStub -
type WhiteListHandlerStub struct {
	RemoveCalled func(keys [][]byte)
	AddCalled    func(keys [][]byte)
}

// Remove -
func (w *WhiteListHandlerStub) Remove(keys [][]byte) {
	if w.RemoveCalled != nil {
		w.RemoveCalled(keys)
	}
}

// Add -
func (w *WhiteListHandlerStub) Add(keys [][]byte) {
	if w.AddCalled != nil {
		w.AddCalled(keys)
	}
}

// IsInterfaceNil -
func (w *WhiteListHandlerStub) IsInterfaceNil() bool {
	return w == nil
}
