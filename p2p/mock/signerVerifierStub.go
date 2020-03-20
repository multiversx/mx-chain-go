package mock

type SignerVerifierStub struct {
	SignCalled      func(message []byte) ([]byte, error)
	VerifyCalled    func(message []byte, sig []byte, pk []byte) error
	PublicKeyCalled func() []byte
}

func (svs *SignerVerifierStub) Sign(message []byte) ([]byte, error) {
	return svs.SignCalled(message)
}

func (svs *SignerVerifierStub) Verify(message []byte, sig []byte, pk []byte) error {
	return svs.VerifyCalled(message, sig, pk)
}

func (svs *SignerVerifierStub) PublicKey() []byte {
	return svs.PublicKeyCalled()
}

func (svs *SignerVerifierStub) IsInterfaceNil() bool {
	return svs == nil
}
