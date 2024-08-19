package shared

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// DelayedBroadcastData is exported to be accessible in delayedBroadcasterMock
type DelayedBroadcastData struct {
	HeaderHash      []byte
	Header          data.HeaderHandler
	MiniBlocksData  map[uint32][]byte
	MiniBlockHashes map[string]map[string]struct{}
	Transactions    map[string][][]byte
	Order           uint32
	PkBytes         []byte
}

// ValidatorHeaderBroadcastData is exported to be accessible in delayedBroadcasterMock
type ValidatorHeaderBroadcastData struct {
	HeaderHash           []byte
	Header               data.HeaderHandler
	MetaMiniBlocksData   map[uint32][]byte
	MetaTransactionsData map[string][][]byte
	Order                uint32
	PkBytes              []byte
}
