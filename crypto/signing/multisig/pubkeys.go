package multisig

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
)

func convertStringsToPubKeys(pubKeys []string, kg crypto.KeyGenerator) ([]crypto.PublicKey, error) {
	// convert pubKeys
	pk := make([]crypto.PublicKey, len(pubKeys))
	for idx, pubKeyStr := range pubKeys {
		if pubKeyStr == "" {
			return nil, crypto.ErrEmptyPubKeyString
		}

		pubKey, err := kg.PublicKeyFromByteArray([]byte(pubKeyStr))
		if err != nil {
			return nil, crypto.ErrInvalidPublicKeyString
		}

		pk[idx] = pubKey
	}
	return pk, nil
}
