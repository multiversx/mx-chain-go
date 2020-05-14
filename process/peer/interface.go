package peer

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

// DataPool indicates the main functionality needed in order to fetch the required blocks from the pool
type DataPool interface {
	Headers() dataRetriever.HeadersPool
	IsInterfaceNil() bool
}

// PeerTypeInfoHandler defines an interface for the PeerTypeInfo
type PeerTypeInfoHandler interface {
	GetPublicKey() string
	GetPeerType() string
	GetShardId() uint32
}
