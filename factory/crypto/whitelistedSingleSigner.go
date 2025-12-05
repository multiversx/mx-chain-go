package crypto

import (
	"encoding/hex"

	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519/singlesig"
)

// ArgsWhiteListedSingleSigner holds the arguments needed to create a whitelisted single signer
type ArgsWhiteListedSingleSigner struct {
	KeyGen                crypto.KeyGenerator
	SingleSigner          *singlesig.Ed25519Signer
	WhitelistedAddressHex string
}

// whitelistedSingleSigner exposes the signing and verification functionalities from the ed25519 signature scheme
type whitelistedSingleSigner struct {
	whitelistedPublicKey crypto.PublicKey
	*singlesig.Ed25519Signer
}

// NewWhiteListEd25519Signer creates a new whitelisted single signer with the provided arguments
func NewWhiteListEd25519Signer(args ArgsWhiteListedSingleSigner) (*whitelistedSingleSigner, error) {
	if args.KeyGen == nil {
		return nil, ErrNilKeyGenerator
	}
	if args.SingleSigner == nil {
		return nil, ErrNilSingleSigner
	}
	if len(args.WhitelistedAddressHex) == 0 {
		return nil, ErrEmptyWhitelistedAddressHex
	}

	whitelistedAddressBytes, err := hex.DecodeString(args.WhitelistedAddressHex)
	if err != nil {
		return nil, err
	}
	whitelistedAddressPublicKey, err := args.KeyGen.PublicKeyFromByteArray(whitelistedAddressBytes)
	if err != nil {
		return nil, err
	}

	return &whitelistedSingleSigner{
		Ed25519Signer:        args.SingleSigner,
		whitelistedPublicKey: whitelistedAddressPublicKey,
	}, nil
}

// Verify verifies a signature using a single signature ed25519 scheme
func (e *whitelistedSingleSigner) Verify(public crypto.PublicKey, msg []byte, sig []byte) error {
	err := e.Ed25519Signer.Verify(public, msg, sig)
	if err == nil {
		return nil
	}

	return e.Ed25519Signer.Verify(e.whitelistedPublicKey, msg, sig)
}
