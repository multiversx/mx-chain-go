package headersCache

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"time"
)

// this structure is only used for sorting
type nonceTimestamp struct {
	nonce     uint64
	timestamp time.Time
}

type headerDetails struct {
	headerHash []byte
	header     data.HeaderHandler
}

type timestampedListOfHeaders struct {
	headers   []headerDetails
	timestamp time.Time
}

func (hld *timestampedListOfHeaders) isEmpty() bool {
	return len(hld.headers) == 0
}

func (hld *timestampedListOfHeaders) removeHeader(index int) {
	hld.headers = append(hld.headers[:index], hld.headers[index+1:]...)
}

func (hld *timestampedListOfHeaders) getHashes() [][]byte {
	hdrsHashes := make([][]byte, 0)
	for _, hdrDetails := range hld.headers {
		hdrsHashes = append(hdrsHashes, hdrDetails.headerHash)
	}

	return hdrsHashes
}

type headerInfo struct {
	headerNonce   uint64
	headerShardId uint32
}
