package mock

// MessageSignVerifierMock -
type MessageSignVerifierMock struct {
}

// Verify -
func (m *MessageSignVerifierMock) Verify(_ []byte, _ []byte, _ []byte) error {
	return nil
}

// IsInterfaceNil -
func (m *MessageSignVerifierMock) IsInterfaceNil() bool {
	return m == nil
}
