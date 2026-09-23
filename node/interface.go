package node

import (
	"io"

	"github.com/multiversx/mx-chain-core-go/core"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// NetworkShardingCollector defines the updating methods used by the network sharding component
// The interface assures that the collected data will be used by the p2p network sharding components
type NetworkShardingCollector interface {
	UpdatePeerIDInfo(pid core.PeerID, pk []byte, shardID uint32)
	PutPeerIdSubType(pid core.PeerID, peerSubType core.P2PPeerSubType)
	GetPeerInfo(pid core.PeerID) core.P2PPeerInfo
	IsInterfaceNil() bool
}

// Throttler can monitor the number of the currently running go routines
type Throttler interface {
	CanProcess() bool
	StartProcessing()
	EndProcessing()
	IsInterfaceNil() bool
}

// HealthService defines the behavior of a service able to keep track of the node's health
type HealthService interface {
	io.Closer
	RegisterComponent(component interface{})
}

type accountHandlerWithDataTrieMigrationStatus interface {
	vmcommon.AccountHandler
	IsDataTrieMigrated() (bool, error)
}
