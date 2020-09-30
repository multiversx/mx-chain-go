package mock

// IsPayableHandlerStub -
type IsPayableHandlerStub struct {
	IsPayableCalled func(address []byte) (bool, error)
}

// IsPayable -
func (i *IsPayableHandlerStub) IsPayable(address []byte) (bool, error) {
	if i.IsPayableCalled != nil {
		return i.IsPayableCalled(address)
	}
	return true, nil
}

// IsInterfaceNil -
func (i *IsPayableHandlerStub) IsInterfaceNil() bool {
	return i == nil
}
