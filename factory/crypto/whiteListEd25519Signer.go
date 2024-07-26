package crypto

import (
	"encoding/hex"

	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519/singlesig"
)

const (
	whitelistedAddressHex    = "285626dbc26423884a493f1f5952c614e92c5aaf2c329690da317a9b49157727"
	xexchangeOwnerAddressHex = "8435c3bc7cec141b87633a87551a766e866255e82cbfa2a4610fea0c88ae5483"
)

// whiteListEd25519Signer exposes the signing and verification functionalities from the ed25519 signature scheme
type whiteListEd25519Signer struct {
	xexchangeOwnerPublicKey crypto.PublicKey
	whitelistedPublicKey    crypto.PublicKey
	singlesig.Ed25519Signer
}

func NewWhiteListEd25519Signer(keyGen crypto.KeyGenerator) (*whiteListEd25519Signer, error) {
	whitelistBytes, err := hex.DecodeString(whitelistedAddressHex)
	if err != nil {
		return nil, err
	}
	pk, err := keyGen.PublicKeyFromByteArray(whitelistBytes)
	if err != nil {
		return nil, err
	}

	xexchangeOwnerPublicKeyBytes, err := hex.DecodeString(xexchangeOwnerAddressHex)
	if err != nil {
		return nil, err
	}
	xexchangeOwnerPublicKey, err := keyGen.PublicKeyFromByteArray(xexchangeOwnerPublicKeyBytes)
	if err != nil {
		return nil, err
	}

	return &whiteListEd25519Signer{
		Ed25519Signer:           singlesig.Ed25519Signer{},
		whitelistedPublicKey:    pk,
		xexchangeOwnerPublicKey: xexchangeOwnerPublicKey,
	}, nil
}

// Verify verifies a signature using a single signature ed25519 scheme
func (e *whiteListEd25519Signer) Verify(public crypto.PublicKey, msg []byte, sig []byte) error {
	if public.Point() == e.xexchangeOwnerPublicKey.Point() {
		return e.Ed25519Signer.Verify(e.whitelistedPublicKey, msg, sig)
	}

	return e.Ed25519Signer.Verify(public, msg, sig)
}
