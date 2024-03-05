package node

import (
	"io"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/debug"
	"github.com/multiversx/mx-chain-go/facade"
	mainFactory "github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/update"
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

// P2PAntifloodHandler defines the behavior of a component able to signal that the system is too busy (or flooded) processing
// p2p messages
type P2PAntifloodHandler interface {
	CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error
	CanProcessMessagesOnTopic(peer core.PeerID, topic string, numMessages uint32, totalSize uint64, sequence []byte) error
	ResetForTopic(topic string)
	SetMaxMessagesForTopic(topic string, maxNum uint32)
	ApplyConsensusSize(size int)
	BlacklistPeer(peer core.PeerID, reason string, duration time.Duration)
	IsInterfaceNil() bool
}

// HardforkTrigger defines the behavior of a hardfork trigger
type HardforkTrigger interface {
	SetExportFactoryHandler(exportFactoryHandler update.ExportFactoryHandler) error
	TriggerReceived(payload []byte, data []byte, pkBytes []byte) (bool, error)
	RecordedTriggerMessage() ([]byte, bool)
	Trigger(epoch uint32, withEarlyEndOfEpoch bool) error
	CreateData() []byte
	AddCloser(closer update.Closer) error
	NotifyTriggerReceivedV2() <-chan struct{}
	IsSelfTrigger() bool
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

// NodeFactory can create a new node
type NodeFactory interface {
	CreateNewNode(opts ...Option) (NodeHandler, error)
	IsInterfaceNil() bool
}

// NodeHandler defines the behavior of a node
type NodeHandler interface {
	facade.NodeHandler
	CreateShardedStores() error
	AddQueryHandler(name string, handler debug.QueryHandler) error
	GetCoreComponents() mainFactory.CoreComponentsHolder
	GetStatusCoreComponents() mainFactory.StatusCoreComponentsHolder
	GetCryptoComponents() mainFactory.CryptoComponentsHolder
	GetConsensusComponents() mainFactory.ConsensusComponentsHolder
	GetBootstrapComponents() mainFactory.BootstrapComponentsHolder
	GetDataComponents() mainFactory.DataComponentsHolder
	GetHeartbeatV2Components() mainFactory.HeartbeatV2ComponentsHolder
	GetNetworkComponents() mainFactory.NetworkComponentsHolder
	GetProcessComponents() mainFactory.ProcessComponentsHolder
	GetStateComponents() mainFactory.StateComponentsHolder
	GetStatusComponents() mainFactory.StatusComponentsHolder
	GetRunTypeComponents() mainFactory.RunTypeComponentsHolder
	Close() error
}
