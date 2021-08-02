package cryptoMocks

import crypto "github.com/ElrondNetwork/elrond-go-crypto"

// MultisignerStub -
type MultisignerStub struct {
	CreateCalled                           func(pubKeys []string, index uint16) (crypto.MultiSigner, error)
	SetAggregatedSigCalled                 func(bytes []byte) error
	VerifyCalled                           func(msg []byte, bitmap []byte) error
	ResetCalled                            func(pubKeys []string, index uint16) error
	CreateSignatureShareCalled             func(msg []byte, bitmap []byte) ([]byte, error)
	StoreSignatureShareCalled              func(index uint16, sig []byte) error
	SignatureShareCalled                   func(index uint16) ([]byte, error)
	VerifySignatureShareCalled             func(index uint16, sig []byte, msg []byte, bitmap []byte) error
	AggregateSigsCalled                    func(bitmap []byte) ([]byte, error)
	CreateAndAddSignatureShareForKeyCalled func(message []byte, privateKey crypto.PrivateKey, pubKeyBytes []byte) ([]byte, error)
}

// Create -
func (mss *MultisignerStub) Create(pubKeys []string, index uint16) (crypto.MultiSigner, error) {
	if mss.CreateCalled != nil {
		return mss.CreateCalled(pubKeys, index)
	}

	return nil, nil
}

// SetAggregatedSig -
func (mss *MultisignerStub) SetAggregatedSig(bytes []byte) error {
	if mss.SetAggregatedSigCalled != nil {
		return mss.SetAggregatedSigCalled(bytes)
	}

	return nil
}

// Verify -
func (mss *MultisignerStub) Verify(msg []byte, bitmap []byte) error {
	if mss.VerifyCalled != nil {
		return mss.VerifyCalled(msg, bitmap)
	}

	return nil
}

// Reset -
func (mss *MultisignerStub) Reset(pubKeys []string, index uint16) error {
	if mss.ResetCalled != nil {
		return mss.ResetCalled(pubKeys, index)
	}

	return nil
}

// CreateSignatureShare -
func (mss *MultisignerStub) CreateSignatureShare(msg []byte, bitmap []byte) ([]byte, error) {
	if mss.CreateSignatureShareCalled != nil {
		return mss.CreateSignatureShareCalled(msg, bitmap)
	}

	return nil, nil
}

// StoreSignatureShare -
func (mss *MultisignerStub) StoreSignatureShare(index uint16, sig []byte) error {
	if mss.StoreSignatureShareCalled != nil {
		return mss.StoreSignatureShareCalled(index, sig)
	}

	return nil
}

// SignatureShare -
func (mss *MultisignerStub) SignatureShare(index uint16) ([]byte, error) {
	if mss.SignatureShareCalled != nil {
		return mss.SignatureShareCalled(index)
	}

	return nil, nil
}

// VerifySignatureShare -
func (mss *MultisignerStub) VerifySignatureShare(index uint16, sig []byte, msg []byte, bitmap []byte) error {
	if mss.VerifySignatureShareCalled != nil {
		return mss.VerifySignatureShareCalled(index, sig, msg, bitmap)
	}

	return nil
}

// AggregateSigs -
func (mss *MultisignerStub) AggregateSigs(bitmap []byte) ([]byte, error) {
	if mss.AggregateSigsCalled != nil {
		return mss.AggregateSigsCalled(bitmap)
	}

	return nil, nil
}

// CreateAndAddSignatureShareForKey -
func (mss *MultisignerStub) CreateAndAddSignatureShareForKey(message []byte, privateKey crypto.PrivateKey, pubKeyBytes []byte) ([]byte, error) {
	if mss.CreateAndAddSignatureShareForKeyCalled != nil {
		return mss.CreateAndAddSignatureShareForKeyCalled(message, privateKey, pubKeyBytes)
	}

	return nil, nil
}

// IsInterfaceNil -
func (mss *MultisignerStub) IsInterfaceNil() bool {
	return mss == nil
}
