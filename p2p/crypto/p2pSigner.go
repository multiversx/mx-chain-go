package crypto

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
)

type p2pSigner struct {
	privateKey *libp2pCrypto.Secp256k1PrivateKey
}

// NewP2PSigner creates a new p2pSigner instance
func NewP2PSigner(privateKey *libp2pCrypto.Secp256k1PrivateKey) (*p2pSigner, error) {
	if privateKey == nil {
		return nil, errNilPrivateKey
	}

	return &p2pSigner{
		privateKey: privateKey,
	}, nil
}

// Sign will sign a payload with the internal private key
func (signer *p2pSigner) Sign(payload []byte) ([]byte, error) {
	return signer.privateKey.Sign(payload)
}

// Verify will check that the (payload, peer ID, signature) tuple is valid or not
func (signer *p2pSigner) Verify(payload []byte, pid core.PeerID, signature []byte) error {
	libp2pPid, err := peer.IDFromBytes(pid.Bytes())
	if err != nil {
		return err
	}

	pubk, err := libp2pPid.ExtractPublicKey()
	if err != nil {
		return fmt.Errorf("cannot extract signing key: %s", err.Error())
	}

	sigOk, err := pubk.Verify(payload, signature)
	if err != nil {
		return err
	}
	if !sigOk {
		return crypto.ErrInvalidSignature
	}

	return nil
}

// SignUsingPrivateKey will sign the payload with provided private key bytes
func (signer *p2pSigner) SignUsingPrivateKey(skBytes []byte, payload []byte) ([]byte, error) {
	sk, err := libp2pCrypto.UnmarshalPrivateKey(skBytes)
	if err != nil {
		return nil, err
	}

	return sk.Sign(payload)
}
