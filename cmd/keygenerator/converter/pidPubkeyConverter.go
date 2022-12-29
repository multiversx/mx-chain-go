package converter

import (
	"encoding/hex"
	"runtime/debug"

	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/secp256k1"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/p2p/factory"
)

var log = logger.GetOrCreate("cmd/keygenerator/converter")

type pidPubkeyConverter struct {
	keyGen crypto.KeyGenerator
}

// NewPidPubkeyConverter creates a new instance of a public key converter that can handle conversions involving core.PeerID string representations
func NewPidPubkeyConverter() *pidPubkeyConverter {
	return &pidPubkeyConverter{
		keyGen: signing.NewKeyGenerator(secp256k1.NewSecp256k1()),
	}
}

// Decode decodes the string as its representation in bytes
func (converter *pidPubkeyConverter) Decode(_ string) ([]byte, error) {
	return nil, errNotImplemented
}

// Encode encodes a byte array in its string representation
func (converter *pidPubkeyConverter) Encode(pkBytes []byte) string {
	pidString, err := converter.encode(pkBytes)
	if err != nil {
		log.Warn("pidPubkeyConverter.Encode encode",
			"hex buff", hex.EncodeToString(pkBytes),
			"error", err,
			"stack trace", string(debug.Stack()),
		)

		return ""
	}

	return pidString
}

func (converter *pidPubkeyConverter) encode(pkBytes []byte) (string, error) {
	pk, err := converter.keyGen.PublicKeyFromByteArray(pkBytes)
	if err != nil {
		return "", err
	}

	pid, err := factory.ConvertPublicKeyToPeerID(pk)
	if err != nil {
		return "", err
	}

	return pid.Pretty(), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (converter *pidPubkeyConverter) IsInterfaceNil() bool {
	return converter == nil
}
