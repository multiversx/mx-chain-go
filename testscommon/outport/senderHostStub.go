package outport

import "github.com/multiversx/mx-chain-communication-go/websocket"

// SenderHostStub -
type SenderHostStub struct {
	SendCalled              func(payload []byte, topic string) error
	SetPayloadHandlerCalled func(handler websocket.PayloadHandler) error
	CloseCalled             func() error
}

// SetPayloadHandler -
func (s *SenderHostStub) SetPayloadHandler(handler websocket.PayloadHandler) error {
	if s.SetPayloadHandlerCalled != nil {
		return s.SetPayloadHandlerCalled(handler)
	}

	return nil
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
