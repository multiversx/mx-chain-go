package outport

// SenderHostStub -
type SenderHostStub struct {
	SendCalled  func(payload []byte, topic string) error
	CloseCalled func() error
}

// Send -
func (s *SenderHostStub) Send(payload []byte, topic string) error {
	if s.SendCalled != nil {
		return s.SendCalled(payload, topic)
	}
	return nil
}

// Close -
func (s *SenderHostStub) Close() error {
	if s.CloseCalled() != nil {
		return s.CloseCalled()
	}
	return nil
}

// IsInterfaceNil -
func (s *SenderHostStub) IsInterfaceNil() bool {
	return s == nil
}
