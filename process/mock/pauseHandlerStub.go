package mock

// PauseHandlerStub -
type PauseHandlerStub struct {
	IsPausedCalled func(token []byte) bool
}

// IsPaused -
func (p *PauseHandlerStub) IsPaused(token []byte) bool {
	if p.IsPausedCalled != nil {
		return p.IsPausedCalled(token)
	}
	return false
}

// IsInterfaceNil -
func (p *PauseHandlerStub) IsInterfaceNil() bool {
	return p == nil
}
