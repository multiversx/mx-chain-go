package sync

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/storage"
)

// blockBootstrapper is the interface needed by base sync to deal with shards and meta nodes while they bootstrap
type blockBootstrapper interface {
	getCurrHeader() (data.HeaderHandler, error)
	getPrevHeader(data.HeaderHandler, storage.Storer) (data.HeaderHandler, error)
	getBlockBody(headerHandler data.HeaderHandler) (data.BodyHandler, error)
	getHeaderWithHashRequestingIfMissing(hash []byte) (data.HeaderHandler, error)
	getHeaderWithNonceRequestingIfMissing(nonce uint64) (data.HeaderHandler, []byte, error)
	getBlockBodyRequestingIfMissing(headerHandler data.HeaderHandler) (data.BodyHandler, error)
	isForkTriggeredByMeta() bool
	requestHeaderByNonce(nonce uint64)
	requestProofByNonce(nonce uint64)
}

// syncStarter defines the behavior of component that can start sync-ing blocks
type syncStarter interface {
	SyncBlock(ctx context.Context) error
}

// settlementChecker answers whether a block at the reconcile nonce is settled, per the settlement
// authority of the chain the node belongs to
type settlementChecker interface {
	prepareInclusionScan(scanCursor uint64) (scanFrom uint64, scanTo uint64, nextCursor uint64)
	isSettled(nonce uint64, headerHash []byte, scanFrom uint64, scanTo uint64) bool
	deadCrossNotarizedMeta() (data.HeaderHandler, []byte, bool)
}

// epochStartTriggerDisarmer reverts a trigger activation armed by a dead epoch start meta block
type epochStartTriggerDisarmer interface {
	DisarmDeadEpochStartActivation(epoch uint32, deadEpochStartHash []byte) bool
}

// forkDetector is the interface needed by base fork detector to deal with shards and meta nodes
type forkDetector interface {
	computeFinalCheckpoint()
}
