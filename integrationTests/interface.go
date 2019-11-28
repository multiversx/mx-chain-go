package integrationTests

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

// TestBootstrapper extends the Bootstrapper interface with some functions intended to be used only in tests
// as it simplifies the reproduction of edge cases
type TestBootstrapper interface {
	process.Bootstrapper
	RollBack(revertUsingForkNonce bool) error
	SetProbableHighestNonce(nonce uint64)
}

// NetworkShardingUpdater defines the updating methods used by the network sharding component
type NetworkShardingUpdater interface {
	ByID(pid p2p.PeerID) (shardId uint32)
	UpdatePeerIdPublicKey(pid p2p.PeerID, pk []byte)
	UpdatePublicKeyShardId(pk []byte, shardId uint32)
	UpdatePeerIdShardId(pid p2p.PeerID, shardId uint32)
	IsInterfaceNil() bool
}
