package processor

import nodesCoord "github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"

// NodesCoordinator defines the operations for a struct that is able to determine if a key is a validator or not
type NodesCoordinator interface {
	GetOwnPublicKey() []byte
	GetValidatorWithPublicKey(publicKey []byte) (nodesCoord.Validator, uint32, error)
	IsInterfaceNil() bool
}
