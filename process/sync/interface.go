package sync

import (
	"context"
	"github.com/multiversx/mx-chain-go/process"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/storage"
)

// blockBootstrapper is the interface needed by base sync to deal with shards and meta nodes while they bootstrap
type blockBootstrapper interface {
	getCurrHeader() (data.HeaderHandler, error)
	getPrevHeader(data.HeaderHandler, storage.Storer) (data.HeaderHandler, error)
	getBlockBody(headerHandler data.HeaderHandler) (data.BodyHandler, error)
	getHeaderWithHashRequestingIfMissing(hash []byte) (data.HeaderHandler, error)
	getHeaderWithNonceRequestingIfMissing(nonce uint64) (data.HeaderHandler, error)
	haveHeaderInPoolWithNonce(nonce uint64) bool
	getBlockBodyRequestingIfMissing(headerHandler data.HeaderHandler) (data.BodyHandler, error)
	isForkTriggeredByMeta() bool
	requestHeaderByNonce(nonce uint64)
}

// syncStarter defines the behavior of component that can start sync-ing blocks
type syncStarter interface {
	SyncBlock(ctx context.Context) error
}

// forkDetector is the interface needed by base fork detector to deal with shards and meta nodes
type forkDetector interface {
	computeFinalCheckpoint()
}

// ForkDetectorCreator is the interface needed by base fork detector to create fork detector
type ForkDetectorCreator interface {
	CreateForkDetector(args ForkDetectorFactoryArgs) (process.ForkDetector, error)
	IsInterfaceNil() bool
}
