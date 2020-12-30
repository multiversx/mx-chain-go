package headersCache

import (
	"bytes"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
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
	items     []headerDetails
	timestamp time.Time
}

func (listOfHeaders *timestampedListOfHeaders) isEmpty() bool {
	return len(listOfHeaders.items) == 0
}

func (listOfHeaders *timestampedListOfHeaders) removeHeader(index int) {
	listOfHeaders.items = append(listOfHeaders.items[:index], listOfHeaders.items[index+1:]...)
}

func (listOfHeaders *timestampedListOfHeaders) getHashes() [][]byte {
	hashes := make([][]byte, 0)
	for _, header := range listOfHeaders.items {
		hashes = append(hashes, header.headerHash)
	}

	return hashes
}

func (listOfHeaders *timestampedListOfHeaders) findHeaderByHash(hash []byte) (data.HeaderHandler, bool) {
	for _, header := range listOfHeaders.items {
		if bytes.Equal(hash, header.headerHash) {
			return header.header, true
		}
	}
	return nil, false
}

type headerInfo struct {
	headerNonce   uint64
	headerShardId uint32
}
