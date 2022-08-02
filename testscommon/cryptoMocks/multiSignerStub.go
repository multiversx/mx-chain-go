package cryptoMocks

// MultiSignerStub implements crypto multisigner
type MultiSignerStub struct {
	VerifyAggregatedSigCalled  func(pubKeysSigners [][]byte, message []byte, aggSig []byte) error
	CreateSignatureShareCalled func(privateKeyBytes []byte, message []byte) ([]byte, error)
	VerifySignatureShareCalled func(publicKey []byte, message []byte, sig []byte) error
	AggregateSigsCalled        func(pubKeysSigners [][]byte, signatures [][]byte) ([]byte, error)
}

// VerifyAggregatedSig -
func (stub *MultiSignerStub) VerifyAggregatedSig(pubKeysSigners [][]byte, message []byte, aggSig []byte) error {
	if stub.VerifyAggregatedSigCalled != nil {
		return stub.VerifyAggregatedSigCalled(pubKeysSigners, message, aggSig)
	}

	return nil
}

// CreateSignatureShare -
func (stub *MultiSignerStub) CreateSignatureShare(privateKeyBytes []byte, message []byte) ([]byte, error) {
	if stub.CreateSignatureShareCalled != nil {
		return stub.CreateSignatureShareCalled(privateKeyBytes, message)
	}

	return nil, nil
}

// VerifySignatureShare -
func (stub *MultiSignerStub) VerifySignatureShare(publicKey []byte, message []byte, sig []byte) error {
	if stub.VerifySignatureShareCalled != nil {
		return stub.VerifySignatureShareCalled(publicKey, message, sig)
	}

	return nil
}

// AggregateSigs -
func (stub *MultiSignerStub) AggregateSigs(pubKeysSigners [][]byte, signatures [][]byte) ([]byte, error) {
	if stub.AggregateSigsCalled != nil {
		return stub.AggregateSigsCalled(pubKeysSigners, signatures)
	}

	return nil, nil
}

// IsInterfaceNil -
func (stub *MultiSignerStub) IsInterfaceNil() bool {
	return stub == nil
}
