package heartbeat

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
)

// NodesCoordinator defines the behavior of a struct able to do validator selection
type NodesCoordinator interface {
	GetValidatorWithPublicKey(publicKey []byte) (validator nodesCoordinator.Validator, shardId uint32, err error)
	IsInterfaceNil() bool
}

// SignaturesHandler defines the behavior of a struct able to handle signatures
type SignaturesHandler interface {
	Verify(payload []byte, pid core.PeerID, signature []byte) error
	IsInterfaceNil() bool
}
