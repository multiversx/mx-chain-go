package mock

type MessageSignVerifierMock struct {
	VerifyCalled func(message []byte, signedMessage []byte, pubKey []byte) error
}

func (m *MessageSignVerifierMock) Verify(message []byte, signedMessage []byte, pubKey []byte) error {
	if m.VerifyCalled != nil {
		return m.VerifyCalled(message, signedMessage, pubKey)
	}
	return nil
}

func (m *MessageSignVerifierMock) IsInterfaceNil() bool {
	return m == nil
}
