package mock

// SignatureHandlerStub implements SignatureHandler interface
type SignatureHandlerStub struct {
	ResetCalled                            func(pubKeys []string) error
	CreateSignatureShareForPublicKeyCalled func(message []byte, index uint16, epoch uint32, publicKeyBytes []byte) ([]byte, error)
	CreateSignatureForPublicKeyCalled      func(message []byte, publicKeyBytes []byte) ([]byte, error)
	StoreSignatureShareCalled              func(index uint16, sig []byte) error
	SignatureShareCalled                   func(index uint16) ([]byte, error)
	VerifySignatureShareCalled             func(index uint16, sig []byte, msg []byte, epoch uint32) error
	AggregateSigsCalled                    func(bitmap []byte, epoch uint32) ([]byte, error)
	SetAggregatedSigCalled                 func(_ []byte) error
	VerifyCalled                           func(msg []byte, bitmap []byte, epoch uint32) error
}

// Reset -
func (stub *SignatureHandlerStub) Reset(pubKeys []string) error {
	if stub.ResetCalled != nil {
		return stub.ResetCalled(pubKeys)
	}

	return nil
}

// CreateSignatureShareForPublicKey -
func (stub *SignatureHandlerStub) CreateSignatureShareForPublicKey(message []byte, index uint16, epoch uint32, publicKeyBytes []byte) ([]byte, error) {
	if stub.CreateSignatureShareForPublicKeyCalled != nil {
		return stub.CreateSignatureShareForPublicKeyCalled(message, index, epoch, publicKeyBytes)
	}

	return make([]byte, 0), nil
}

// CreateSignatureForPublicKey -
func (stub *SignatureHandlerStub) CreateSignatureForPublicKey(message []byte, publicKeyBytes []byte) ([]byte, error) {
	if stub.CreateSignatureForPublicKeyCalled != nil {
		return stub.CreateSignatureForPublicKeyCalled(message, publicKeyBytes)
	}

	return make([]byte, 0), nil
}

// StoreSignatureShare -
func (stub *SignatureHandlerStub) StoreSignatureShare(index uint16, sig []byte) error {
	if stub.StoreSignatureShareCalled != nil {
		return stub.StoreSignatureShareCalled(index, sig)
	}

	return nil
}

// SignatureShare -
func (stub *SignatureHandlerStub) SignatureShare(index uint16) ([]byte, error) {
	if stub.SignatureShareCalled != nil {
		return stub.SignatureShareCalled(index)
	}

	return []byte("sigShare"), nil
}

// VerifySignatureShare -
func (stub *SignatureHandlerStub) VerifySignatureShare(index uint16, sig []byte, msg []byte, epoch uint32) error {
	if stub.VerifySignatureShareCalled != nil {
		return stub.VerifySignatureShareCalled(index, sig, msg, epoch)
	}

	return nil
}

// AggregateSigs -
func (stub *SignatureHandlerStub) AggregateSigs(bitmap []byte, epoch uint32) ([]byte, error) {
	if stub.AggregateSigsCalled != nil {
		return stub.AggregateSigsCalled(bitmap, epoch)
	}

	return []byte("aggSigs"), nil
}

// SetAggregatedSig -
func (stub *SignatureHandlerStub) SetAggregatedSig(sig []byte) error {
	if stub.SetAggregatedSigCalled != nil {
		return stub.SetAggregatedSigCalled(sig)
	}

	return nil
}

// Verify -
func (stub *SignatureHandlerStub) Verify(msg []byte, bitmap []byte, epoch uint32) error {
	if stub.VerifyCalled != nil {
		return stub.VerifyCalled(msg, bitmap, epoch)
	}

	return nil
}

// IsInterfaceNil -
func (stub *SignatureHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
