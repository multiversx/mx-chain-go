package p2pmocks

import (
	"github.com/multiversx/mx-chain-core-go/core"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
)

// P2PKeyConverterStub -
type P2PKeyConverterStub struct {
	ConvertPeerIDToPublicKeyCalled func(keyGen crypto.KeyGenerator, pid core.PeerID) (crypto.PublicKey, error)
	ConvertPublicKeyToPeerIDCalled func(pk crypto.PublicKey) (core.PeerID, error)
}

// ConvertPeerIDToPublicKey -
func (stub *P2PKeyConverterStub) ConvertPeerIDToPublicKey(keyGen crypto.KeyGenerator, pid core.PeerID) (crypto.PublicKey, error) {
	if stub.ConvertPeerIDToPublicKeyCalled != nil {
		return stub.ConvertPeerIDToPublicKeyCalled(keyGen, pid)
	}

	return &cryptoMocks.PublicKeyStub{}, nil
}

// ConvertPublicKeyToPeerID -
func (stub *P2PKeyConverterStub) ConvertPublicKeyToPeerID(pk crypto.PublicKey) (core.PeerID, error) {
	if stub.ConvertPublicKeyToPeerIDCalled != nil {
		return stub.ConvertPublicKeyToPeerIDCalled(pk)
	}

	return "", nil
}

// IsInterfaceNil -
func (stub *P2PKeyConverterStub) IsInterfaceNil() bool {
	return stub == nil
}
