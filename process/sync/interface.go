package sync

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// storageBootstrapper is the main interface for bootstrap from storage execution engine
type storageBootstrapper interface {
	getHeader(shardId uint32, nonce uint64) (data.HeaderHandler, []byte, error)
	getBlockBody(data.HeaderHandler) (data.BodyHandler, error)
	removeBlockBody(nonce uint64, blockUnit dataRetriever.UnitType, hdrNonceHashDataUnit dataRetriever.UnitType) error
	getNonceWithLastNotarized(currrentNonce uint64) (startNonce uint64, finalNotarized map[uint32]uint64, lastNotarized map[uint32]uint64)
	applyNotarizedBlocks(finalNotarized map[uint32]uint64, lastNotarized map[uint32]uint64) error
	cleanupNotarizedStorage(lastNotarized map[uint32]uint64)
	IsInterfaceNil() bool
}

// blockBootstrapper is the interface needed by base sync to deal with shards and meta nodes while they bootstrap
type blockBootstrapper interface {
	getCurrHeader() (data.HeaderHandler, error)
	getPrevHeader(data.HeaderHandler, storage.Storer) (data.HeaderHandler, error)
	getBlockBody(headerHandler data.HeaderHandler) (data.BodyHandler, error)
	getHeaderWithHashRequestingIfMissing(hash []byte) (data.HeaderHandler, error)
	getHeaderWithNonceRequestingIfMissing(nonce uint64) (data.HeaderHandler, error)
	haveHeaderInPoolWithNonce(nonce uint64) bool
	getBlockBodyRequestingIfMissing(headerHandler data.HeaderHandler) (data.BodyHandler, error)
	getCurrHeaderHash() []byte
	isForkTriggeredByMeta() bool
}
