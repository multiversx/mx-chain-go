package mock

// PayableHandlerStub -
type PayableHandlerStub struct {
	IsPayableCalled func(sndAddress []byte, recvAdress []byte) (bool, error)
}

// IsPayable -
func (p *PayableHandlerStub) IsPayable(sndAddress []byte, recvAddress []byte) (bool, error) {
	if p.IsPayableCalled != nil {
		return p.IsPayableCalled(sndAddress, recvAddress)
	}
	return true, nil
}

// IsInterfaceNil -
func (p *PayableHandlerStub) IsInterfaceNil() bool {
	return p == nil
}
