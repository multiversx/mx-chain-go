package mock

// SignatureHandlerStub implements SignatureHandler interface
type SignatureHandlerStub struct {
	ResetCalled                 func(pubKeys []string) error
	SetMultiSignerByEpochCalled func(epoch uint32) error
	CreateSignatureShareCalled  func(msg []byte, index uint16) ([]byte, error)
	StoreSignatureShareCalled   func(index uint16, sig []byte) error
	SignatureShareCalled        func(index uint16) ([]byte, error)
	VerifySignatureShareCalled  func(index uint16, sig []byte, msg []byte) error
	AggregateSigsCalled         func(bitmap []byte) ([]byte, error)
	SetAggregatedSigCalled      func(_ []byte) error
	VerifyCalled                func(msg []byte, bitmap []byte) error
}

// SetMultiSignerByEpoch -
func (stub *SignatureHandlerStub) SetMultiSignerByEpoch(epoch uint32) error {
	if stub.SetMultiSignerByEpochCalled != nil {
		return stub.SetMultiSignerByEpochCalled(epoch)
	}

	return nil
}

// Reset -
func (stub *SignatureHandlerStub) Reset(pubKeys []string) error {
	if stub.ResetCalled != nil {
		return stub.ResetCalled(pubKeys)
	}

	return nil
}

// CreateSignatureShare -
func (stub *SignatureHandlerStub) CreateSignatureShare(msg []byte, index uint16) ([]byte, error) {
	if stub.CreateSignatureShareCalled != nil {
		return stub.CreateSignatureShareCalled(msg, index)
	}

	return []byte("sigShare"), nil
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
func (stub *SignatureHandlerStub) VerifySignatureShare(index uint16, sig []byte, msg []byte) error {
	if stub.VerifySignatureShareCalled != nil {
		return stub.VerifySignatureShareCalled(index, sig, msg)
	}

	return nil
}

// AggregateSigs -
func (stub *SignatureHandlerStub) AggregateSigs(bitmap []byte) ([]byte, error) {
	if stub.AggregateSigsCalled != nil {
		return stub.AggregateSigsCalled(bitmap)
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
func (stub *SignatureHandlerStub) Verify(msg []byte, bitmap []byte) error {
	if stub.VerifyCalled != nil {
		return stub.VerifyCalled(msg, bitmap)
	}

	return nil
}

// IsInterfaceNil -
func (stub *SignatureHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
