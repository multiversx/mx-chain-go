package processor

import "github.com/ElrondNetwork/elrond-go-core/core"

// NodesCoordinator defines the operations for a struct that is able to determine if a key is a validator or not
type NodesCoordinator interface {
	GetOwnPublicKey() []byte
	GetValidatorWithPublicKey(publicKey []byte) (core.Validator, uint32, error)
	IsInterfaceNil() bool
}
