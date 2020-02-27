package mock

type MessageSignVerifierMock struct {
}

func (m *MessageSignVerifierMock) Verify(message []byte, signedMessage []byte, pubKey []byte) error {
	return nil
}

func (m *MessageSignVerifierMock) IsInterfaceNil() bool {
	return m == nil
}
