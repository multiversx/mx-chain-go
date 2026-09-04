package cryptoMocks

import crypto "github.com/multiversx/mx-chain-crypto-go"

// MultiSignerStub implements crypto multisigner
type MultiSignerStub struct {
	CreateSignatureShareCalled   func(privateKeyBytes []byte, message []byte) ([]byte, error)
	CreateSignatureShareV2Called func(privateKeyBytes crypto.PrivateKey, message []byte) ([]byte, error)
	VerifySignatureShareCalled   func(publicKey []byte, message []byte, sig []byte) error
	VerifySignatureShareV2Called func(publicKey crypto.PublicKey, message []byte, sig []byte) error
	AggregateSigsCalled          func(pubKeysSigners [][]byte, signatures [][]byte) ([]byte, error)
	AggregateSigsV2Called        func(pubKeysSigners []crypto.PublicKey, signatures [][]byte) ([]byte, error)
	VerifyAggregatedSigCalled    func(pubKeysSigners [][]byte, message []byte, aggSig []byte) error
	VerifyAggregatedSigV2Called  func(pubKeysSigners []crypto.PublicKey, message []byte, aggSig []byte) error
}

// VerifyAggregatedSig -
func (stub *MultiSignerStub) VerifyAggregatedSig(pubKeysSigners [][]byte, message []byte, aggSig []byte) error {
	if stub.VerifyAggregatedSigCalled != nil {
		return stub.VerifyAggregatedSigCalled(pubKeysSigners, message, aggSig)
	}

	return nil
}

// VerifyAggregatedSigV2 -
func (stub *MultiSignerStub) VerifyAggregatedSigV2(pubKeysSigners []crypto.PublicKey, message []byte, aggSig []byte) error {
	if stub.VerifyAggregatedSigV2Called != nil {
		return stub.VerifyAggregatedSigV2Called(pubKeysSigners, message, aggSig)
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

// CreateSignatureShareV2 -
func (stub *MultiSignerStub) CreateSignatureShareV2(privateKey crypto.PrivateKey, message []byte) ([]byte, error) {
	if stub.CreateSignatureShareV2Called != nil {
		return stub.CreateSignatureShareV2Called(privateKey, message)
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

// VerifySignatureShareV2 -
func (stub *MultiSignerStub) VerifySignatureShareV2(publicKey crypto.PublicKey, message []byte, sig []byte) error {
	if stub.VerifySignatureShareV2Called != nil {
		return stub.VerifySignatureShareV2Called(publicKey, message, sig)
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

// AggregateSigsV2 -
func (stub *MultiSignerStub) AggregateSigsV2(pubKeys []crypto.PublicKey, signatures [][]byte) ([]byte, error) {
	if stub.AggregateSigsV2Called != nil {
		return stub.AggregateSigsV2Called(pubKeys, signatures)
	}

	return nil, nil
}

// IsInterfaceNil -
func (stub *MultiSignerStub) IsInterfaceNil() bool {
	return stub == nil
}
