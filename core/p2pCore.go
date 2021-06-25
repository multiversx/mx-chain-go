package core

// P2PPeerType defines the type of a p2p peer
type P2PPeerType int

// String returns the string-ified version of P2PPeerType
func (pt P2PPeerType) String() string {
	switch pt {
	case ValidatorPeer:
		return "validator"
	case ObserverPeer:
		return "observer"
	default:
		return "unknown"
	}
}

const (
	// UnknownPeer defines a peer that is unknown (did not advertise data in any way)
	UnknownPeer P2PPeerType = iota
	// ValidatorPeer means that the peer is a validator
	ValidatorPeer
	// ObserverPeer means that the peer is an observer
	ObserverPeer
)

// P2PPeerSubType defines the subtype of peer (e.g. FullArchive)
type P2PPeerSubType uint32

const (
	// RegularPeer
	RegularPeer P2PPeerSubType = iota
	// FullHistoryObserver is a node that syncs the entire history of its shard
	FullHistoryObserver
)

// String returns the string-ified version of P2PPeerSubType
func (pst P2PPeerSubType) String() string {
	switch pst {
	case RegularPeer:
		return "regular"
	case FullHistoryObserver:
		return "fullArchive"
	default:
		return "unknown"
	}
}

// P2PPeerInfo represents a peer info structure
type P2PPeerInfo struct {
	PeerType    P2PPeerType
	PeerSubType P2PPeerSubType
	ShardID     uint32
	PkBytes     []byte
}

// QueryP2PPeerInfo represents a DTO used in exporting p2p peer info after a query
type QueryP2PPeerInfo struct {
	IsBlacklisted bool     `json:"isblacklisted"`
	Pid           string   `json:"pid"`
	Pk            string   `json:"pk"`
	PeerType      string   `json:"peertype"`
	PeerSubType   string   `json:"peersubtype"`
	Addresses     []string `json:"addresses"`
}

// PeerTopicType represents the type of a peer in regards to the topic it is used
type PeerTopicType string

// String returns a string version of the peer topic type
func (pt PeerTopicType) String() string {
	return string(pt)
}

const (
	// IntraShardPeer represents the identifier for intra shard peers to be used in intra shard topics
	IntraShardPeer PeerTopicType = "intra peer"

	// CrossShardPeer represents the identifier for intra shard peers to be used in cross shard topics
	CrossShardPeer PeerTopicType = "cross peer"

	// FullHistoryPeer represents the identifier for intra shard peers to be used in full history topics
	FullHistoryPeer PeerTopicType = "full history peer"
)
