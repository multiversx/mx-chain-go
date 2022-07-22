package cryptoMocks

// MultiSignerNewStub implements crypto multisigner
// TODO: change name
type MultiSignerNewStub struct {
	VerifyAggregatedSigCalled  func(pubKeysSigners [][]byte, message []byte, aggSig []byte) error
	CreateSignatureShareCalled func(privateKeyBytes []byte, message []byte) ([]byte, error)
	VerifySignatureShareCalled func(publicKey []byte, message []byte, sig []byte) error
	AggregateSigsCalled        func(pubKeysSigners [][]byte, signatures [][]byte) ([]byte, error)
}

// VerifyAggregatedSig -
func (stub *MultiSignerNewStub) VerifyAggregatedSig(pubKeysSigners [][]byte, message []byte, aggSig []byte) error {
	if stub.VerifyAggregatedSigCalled != nil {
		return stub.VerifyAggregatedSigCalled(pubKeysSigners, message, aggSig)
	}

	return nil
}

// CreateSignatureShare -
func (stub *MultiSignerNewStub) CreateSignatureShare(privateKeyBytes []byte, message []byte) ([]byte, error) {
	if stub.CreateSignatureShareCalled != nil {
		return stub.CreateSignatureShareCalled(privateKeyBytes, message)
	}

	return nil, nil
}

// VerifySignatureShare -
func (stub *MultiSignerNewStub) VerifySignatureShare(publicKey []byte, message []byte, sig []byte) error {
	if stub.VerifySignatureShareCalled != nil {
		return stub.VerifySignatureShareCalled(publicKey, message, sig)
	}

	return nil
}

// AggregateSigs -
func (stub *MultiSignerNewStub) AggregateSigs(pubKeysSigners [][]byte, signatures [][]byte) ([]byte, error) {
	if stub.AggregateSigsCalled != nil {
		return stub.AggregateSigsCalled(pubKeysSigners, signatures)
	}

	return nil, nil
}

// IsInterfaceNil -
func (stub *MultiSignerNewStub) IsInterfaceNil() bool {
	return stub == nil
}
