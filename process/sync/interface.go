package sync

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

type storageBootstrapper interface {
	getHeader(shardId uint32, nonce uint64) (data.HeaderHandler, []byte, error)
	getBlockBody(data.HeaderHandler) (data.BodyHandler, error)
	removeBlockBody(nonce uint64, blockUnit dataRetriever.UnitType, hdrNonceHashDataUnit dataRetriever.UnitType) error
	getNonceWithLastNotarized(currrentNonce uint64) (startNonce uint64, finalNotarized map[uint32]uint64, lastNotarized map[uint32]uint64)
	applyNotarizedBlocks(finalNotarized map[uint32]uint64, lastNotarized map[uint32]uint64) error
	cleanupNotarizedStorage(lastNotarized map[uint32]uint64)
}
