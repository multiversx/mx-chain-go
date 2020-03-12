package mock

// MessageSignVerifierMock -
type MessageSignVerifierMock struct {
	VerifyCalled func(message []byte, signedMessage []byte, pubKey []byte) error
}

// Verify -
func (m *MessageSignVerifierMock) Verify(message []byte, signedMessage []byte, pubKey []byte) error {
	if m.VerifyCalled != nil {
		return m.VerifyCalled(message, signedMessage, pubKey)
	}
	return nil
}

// IsInterfaceNil -
func (m *MessageSignVerifierMock) IsInterfaceNil() bool {
	return m == nil
}
