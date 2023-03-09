package cryptoMocks

import (
	"github.com/multiversx/mx-chain-core-go/core"
	crypto "github.com/multiversx/mx-chain-crypto-go"
)

// PeerSignatureHandlerStub -
type PeerSignatureHandlerStub struct {
	VerifyPeerSignatureCalled func(pk []byte, pid core.PeerID, signature []byte) error
	GetPeerSignatureCalled    func(key crypto.PrivateKey, pid []byte) ([]byte, error)
}

// VerifyPeerSignature -
func (pshs *PeerSignatureHandlerStub) VerifyPeerSignature(pk []byte, pid core.PeerID, signature []byte) error {
	if pshs.VerifyPeerSignatureCalled != nil {
		return pshs.VerifyPeerSignatureCalled(pk, pid, signature)
	}
	return nil
}

// GetPeerSignature -
func (pshs *PeerSignatureHandlerStub) GetPeerSignature(key crypto.PrivateKey, pid []byte) ([]byte, error) {
	if pshs.GetPeerSignatureCalled != nil {
		return pshs.GetPeerSignatureCalled(key, pid)
	}
	return nil, nil
}

// IsInterfaceNil -
func (pshs *PeerSignatureHandlerStub) IsInterfaceNil() bool {
	return false
}
