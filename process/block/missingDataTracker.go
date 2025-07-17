package block

import (
	"fmt"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
)

const checkMissingDataStep = 10 * time.Millisecond

type missingDataTracker struct {
	mutHeaders     sync.RWMutex
	missingHeaders map[string]struct{}
	mutProofs      sync.RWMutex
	missingProofs  map[string]struct{}
	headersPool    dataRetriever.HeadersPool
	proofsPool     dataRetriever.ProofsPool
}

func newMissingDataTracker(headersPool dataRetriever.HeadersPool, proofsPool dataRetriever.ProofsPool) *missingDataTracker {
	mdt := &missingDataTracker{
		missingHeaders: make(map[string]struct{}),
		missingProofs:  make(map[string]struct{}),
		headersPool:    headersPool,
		proofsPool:     proofsPool,
	}

	mdt.monitorReceivedData()
	return mdt
}

func (tracker *missingDataTracker) addMissingHeader(hash []byte) {
	tracker.mutHeaders.Lock()
	tracker.missingHeaders[string(hash)] = struct{}{}
	tracker.mutHeaders.Unlock()
}

func (tracker *missingDataTracker) addMissingProof(hash []byte) {
	tracker.mutProofs.Lock()
	tracker.missingProofs[string(hash)] = struct{}{}
	tracker.mutProofs.Unlock()
}

func (tracker *missingDataTracker) markHeaderReceived(hash []byte) {
	tracker.mutHeaders.Lock()
	delete(tracker.missingHeaders, string(hash))
	tracker.mutHeaders.Unlock()
}

func (tracker *missingDataTracker) markProofReceived(hash []byte) {
	tracker.mutProofs.Lock()
	delete(tracker.missingProofs, string(hash))
	tracker.mutProofs.Unlock()
}

func (tracker *missingDataTracker) allHeadersReceived() bool {
	tracker.mutHeaders.RLock()
	defer tracker.mutHeaders.RUnlock()

	return len(tracker.missingHeaders) == 0
}

func (tracker *missingDataTracker) allProofsReceived() bool {
	tracker.mutProofs.RLock()
	defer tracker.mutProofs.RUnlock()

	return len(tracker.missingProofs) == 0
}

func (tracker *missingDataTracker) allDataReceived() bool {
	return tracker.allHeadersReceived() && tracker.allProofsReceived()
}

func (tracker *missingDataTracker) receivedProof(proof data.HeaderProofHandler) {
	tracker.markProofReceived(proof.GetHeaderHash())
}

func (tracker *missingDataTracker) receivedHeader(_ data.HeaderHandler, headerHash []byte) {
	tracker.markHeaderReceived(headerHash)
}

func (tracker *missingDataTracker) monitorReceivedData() {
	tracker.headersPool.RegisterHandler(tracker.receivedHeader)
	tracker.proofsPool.RegisterHandler(tracker.receivedProof)
}

func (sp *shardProcessor) requestHeaderIfNeeded(
	shardID uint32,
	headerHash []byte,
	tracker *missingDataTracker,
) {
	_, err := sp.dataPool.Headers().GetHeaderByHash(headerHash)
	if err == nil {
		return
	}

	tracker.addMissingHeader(headerHash)
	if shardID == core.MetachainShardId {
		go sp.requestHandler.RequestMetaHeader(headerHash)
	} else {
		go sp.requestHandler.RequestShardHeader(shardID, headerHash)
	}
}

func (sp *shardProcessor) requestProofIfNeeded(shardID uint32, headerHash []byte, tracker *missingDataTracker) {
	if sp.proofsPool.HasProof(shardID, headerHash) {
		return
	}

	tracker.addMissingProof(headerHash)
	go sp.requestHandler.RequestEquivalentProofByHash(shardID, headerHash)
}

func (sp *shardProcessor) RequestMissingMetaHeadersBlocking(
	tracker *missingDataTracker,
	shardHeader data.ShardHeaderHandler,
	timeout time.Duration,
) error {
	if check.IfNil(shardHeader) {
		return process.ErrNilBlockHeader
	}

	metaBlockHashes := shardHeader.GetMetaBlockHashes()
	for i := 0; i < len(metaBlockHashes); i++ {
		sp.requestHeaderIfNeeded(core.MetachainShardId, metaBlockHashes[i], tracker)
		sp.requestProofIfNeeded(core.MetachainShardId, metaBlockHashes[i], tracker)
	}

	return tracker.waitForMissingData(timeout)
}

// todo: maybe use channels instead of polling
func (tracker *missingDataTracker) waitForMissingData(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		if tracker.allDataReceived() {
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for missing data")
		}
		time.Sleep(checkMissingDataStep)
	}
}
