package heartbeat

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
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
