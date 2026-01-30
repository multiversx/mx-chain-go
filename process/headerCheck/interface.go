package headerCheck

import crypto "github.com/multiversx/mx-chain-crypto-go"

type pubKeysHandler interface {
	GetPubKeysFromBytes(pubKeysBytes [][]byte) ([]crypto.PublicKey, error)
	IsInterfaceNil() bool
}
