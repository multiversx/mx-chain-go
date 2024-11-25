package crypto

import (
	"encoding/hex"

	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519/singlesig"
)

const (
	whitelistedAddressHex    = "285626dbc26423884a493f1f5952c614e92c5aaf2c329690da317a9b49157727" //erd19ptzdk7zvs3csjjf8u04j5kxzn5jck409sefdyx6x9afkjg4wunsfw7rj7
	xexchangeOwnerAddressHex = "8435c3bc7cec141b87633a87551a766e866255e82cbfa2a4610fea0c88ae5483" //erd1ss6u80ruas2phpmr82r42xnkd6rxy40g9jl69frppl4qez9w2jpsqj8x97
	faucetAddressHex         = "1a6b7679fd1f1a81720f26e7384f150efef9588e428d260a1f43f62ace6a5a3c" //erd1rf4hv70arudgzus0ymnnsnc4pml0jkywg2xjvzslg0mz4nn2tg7q7k0t6p
)

// whiteListEd25519Signer exposes the signing and verification functionalities from the ed25519 signature scheme
type whiteListEd25519Signer struct {
	xexchangeOwnerPublicKey crypto.PublicKey
	faucetAddressPublicKey  crypto.PublicKey
	whitelistedPublicKey    crypto.PublicKey
	singlesig.Ed25519Signer
}

func NewWhiteListEd25519Signer(keyGen crypto.KeyGenerator) (*whiteListEd25519Signer, error) {
	whitelistedAddressBytes, err := hex.DecodeString(whitelistedAddressHex)
	if err != nil {
		return nil, err
	}
	whitelistedAddressPublicKey, err := keyGen.PublicKeyFromByteArray(whitelistedAddressBytes)
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

	faucetAddressPublicKeyBytes, err := hex.DecodeString(faucetAddressHex)
	if err != nil {
		return nil, err
	}
	faucetAddressPublicKey, err := keyGen.PublicKeyFromByteArray(faucetAddressPublicKeyBytes)
	if err != nil {
		return nil, err
	}

	return &whiteListEd25519Signer{
		Ed25519Signer:           singlesig.Ed25519Signer{},
		whitelistedPublicKey:    whitelistedAddressPublicKey,
		xexchangeOwnerPublicKey: xexchangeOwnerPublicKey,
		faucetAddressPublicKey:  faucetAddressPublicKey,
	}, nil
}

// Verify verifies a signature using a single signature ed25519 scheme
func (e *whiteListEd25519Signer) Verify(public crypto.PublicKey, msg []byte, sig []byte) error {
	isExchangeOwner, err := public.Point().Equal(e.xexchangeOwnerPublicKey.Point())
	if err != nil {
		return err
	}
	if isExchangeOwner {
		return e.Ed25519Signer.Verify(e.whitelistedPublicKey, msg, sig)
	}

	isFaucetAddress, err := public.Point().Equal(e.faucetAddressPublicKey.Point())
	if err != nil {
		return err
	}
	if isFaucetAddress {
		return e.Ed25519Signer.Verify(e.whitelistedPublicKey, msg, sig)
	}

	return e.Ed25519Signer.Verify(public, msg, sig)
}
