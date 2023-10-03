package converter

import (
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/secp256k1"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/p2p/factory"
)

type pidPubkeyConverter struct {
	keyGen          crypto.KeyGenerator
	p2PKeyConverter p2p.P2PKeyConverter
}

// NewPidPubkeyConverter creates a new instance of a public key converter that can handle conversions involving core.PeerID string representations
func NewPidPubkeyConverter() *pidPubkeyConverter {
	return &pidPubkeyConverter{
		keyGen:          signing.NewKeyGenerator(secp256k1.NewSecp256k1()),
		p2PKeyConverter: factory.NewP2PKeyConverter(),
	}
}

// Decode decodes the string as its representation in bytes
func (converter *pidPubkeyConverter) Decode(_ string) ([]byte, error) {
	return nil, errNotImplemented
}

// Encode encodes a byte array in its string representation
func (converter *pidPubkeyConverter) Encode(pkBytes []byte) (string, error) {
	pidString, err := converter.encode(pkBytes)
	if err != nil {
		return "", err
	}

	return pidString, nil
}

func (converter *pidPubkeyConverter) encode(pkBytes []byte) (string, error) {
	pk, err := converter.keyGen.PublicKeyFromByteArray(pkBytes)
	if err != nil {
		return "", err
	}

	pid, err := converter.p2PKeyConverter.ConvertPublicKeyToPeerID(pk)
	if err != nil {
		return "", err
	}

	return pid.Pretty(), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (converter *pidPubkeyConverter) IsInterfaceNil() bool {
	return converter == nil
}
