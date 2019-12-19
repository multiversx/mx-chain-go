package track

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

// blockTracker is the interface needed by base block track to deal with shards and meta nodes while they track blocks
type blockTracker interface {
	getSelfHeaders(headerHandler data.HeaderHandler) []*headerInfo
	computeLongestSelfChain() (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte)
}
