package mock

// PayableHandlerStub -
type PayableHandlerStub struct {
	IsPayableCalled func(address []byte) (bool, error)
}

// IsPayable -
func (p *PayableHandlerStub) IsPayable(address []byte) (bool, error) {
	if p.IsPayableCalled != nil {
		return p.IsPayableCalled(address)
	}
	return true, nil
}

// IsInterfaceNil -
func (p *PayableHandlerStub) IsInterfaceNil() bool {
	return p == nil
}
