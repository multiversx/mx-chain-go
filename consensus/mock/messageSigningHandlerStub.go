package mock

import "github.com/multiversx/mx-chain-go/p2p"

// MessageSigningHandlerStub -
type MessageSigningHandlerStub struct {
	VerifyCalled      func(message p2p.MessageP2P) error
	SerializeCalled   func(messages []p2p.MessageP2P) ([]byte, error)
	DeserializeCalled func(messagesBytes []byte) ([]p2p.MessageP2P, error)
}

// Verify -
func (stub *MessageSigningHandlerStub) Verify(message p2p.MessageP2P) error {
	if stub.VerifyCalled != nil {
		return stub.VerifyCalled(message)
	}

	return nil
}

// Serialize -
func (stub *MessageSigningHandlerStub) Serialize(messages []p2p.MessageP2P) ([]byte, error) {
	if stub.SerializeCalled != nil {
		return stub.SerializeCalled(messages)
	}

	return nil, nil
}

// Deserialize -
func (stub *MessageSigningHandlerStub) Deserialize(messagesBytes []byte) ([]p2p.MessageP2P, error) {
	if stub.DeserializeCalled != nil {
		return stub.DeserializeCalled(messagesBytes)
	}

	return nil, nil
}

// IsInterfaceNil -
func (stub *MessageSigningHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
