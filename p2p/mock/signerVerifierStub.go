package mock

// SignerVerifierStub -
type SignerVerifierStub struct {
	SignCalled      func(message []byte) ([]byte, error)
	VerifyCalled    func(message []byte, sig []byte, pk []byte) error
	PublicKeyCalled func() []byte
}

// Sign -
func (svs *SignerVerifierStub) Sign(message []byte) ([]byte, error) {
	return svs.SignCalled(message)
}

// Verify -
func (svs *SignerVerifierStub) Verify(message []byte, sig []byte, pk []byte) error {
	return svs.VerifyCalled(message, sig, pk)
}

// PublicKey -
func (svs *SignerVerifierStub) PublicKey() []byte {
	return svs.PublicKeyCalled()
}

// IsInterfaceNil -
func (svs *SignerVerifierStub) IsInterfaceNil() bool {
	return svs == nil
}
