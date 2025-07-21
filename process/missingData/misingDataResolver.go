package missingData

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

var errTimeoutWaitingForMissingData = fmt.Errorf("timeout waiting for missing data")

type missingDataResolver struct {
	mutHeaders     sync.RWMutex
	missingHeaders map[string]struct{}
	mutProofs      sync.RWMutex
	missingProofs  map[string]struct{}
	headersPool    dataRetriever.HeadersPool
	proofsPool     dataRetriever.ProofsPool
	requestHandler process.RequestHandler
}

func NewMissingDataResolver(headersPool dataRetriever.HeadersPool, proofsPool dataRetriever.ProofsPool, requestHandler process.RequestHandler) (*missingDataResolver, error) {
	if check.IfNil(headersPool) {
		return nil, process.ErrNilHeadersDataPool
	}
	if check.IfNil(proofsPool) {
		return nil, process.ErrNilProofsPool
	}
	if check.IfNil(requestHandler) {
		return nil, process.ErrNilRequestHandler
	}

	mdt := &missingDataResolver{
		missingHeaders: make(map[string]struct{}),
		missingProofs:  make(map[string]struct{}),
		headersPool:    headersPool,
		proofsPool:     proofsPool,
		requestHandler: requestHandler,
	}

	mdt.monitorReceivedData()
	return mdt, nil
}

// RequestMissingMetaHeadersBlocking requests the missing meta headers and proofs for the given shard header.
func (mdr *missingDataResolver) RequestMissingMetaHeadersBlocking(
	shardHeader data.ShardHeaderHandler,
	timeout time.Duration,
) error {
	if check.IfNil(shardHeader) {
		return process.ErrNilBlockHeader
	}

	metaBlockHashes := shardHeader.GetMetaBlockHashes()
	for i := 0; i < len(metaBlockHashes); i++ {
		mdr.requestHeaderIfNeeded(core.MetachainShardId, metaBlockHashes[i])
		mdr.requestProofIfNeeded(core.MetachainShardId, metaBlockHashes[i])
	}

	return mdr.waitForMissingData(timeout)
}

func (mdr *missingDataResolver) addMissingHeader(hash []byte) bool {
	mdr.mutHeaders.Lock()
	mdr.missingHeaders[string(hash)] = struct{}{}
	mdr.mutHeaders.Unlock()

	// avoid missing notifications if the header just arrived
	_, err := mdr.headersPool.GetHeaderByHash(hash)
	if err == nil {
		mdr.mutHeaders.Lock()
		delete(mdr.missingHeaders, string(hash))
		mdr.mutHeaders.Unlock()
	}

	return err != nil
}

func (mdr *missingDataResolver) addMissingProof(shardID uint32, hash []byte) bool {
	mdr.mutProofs.Lock()
	mdr.missingProofs[string(hash)] = struct{}{}
	mdr.mutProofs.Unlock()

	// avoid missing notifications if the proof just arrived
	hasProof := mdr.proofsPool.HasProof(shardID, hash)
	if hasProof {
		mdr.mutProofs.Lock()
		delete(mdr.missingProofs, string(hash))
		mdr.mutProofs.Unlock()
	}

	return !hasProof
}

func (mdr *missingDataResolver) markHeaderReceived(hash []byte) {
	mdr.mutHeaders.Lock()
	delete(mdr.missingHeaders, string(hash))
	mdr.mutHeaders.Unlock()
}

func (mdr *missingDataResolver) markProofReceived(hash []byte) {
	mdr.mutProofs.Lock()
	delete(mdr.missingProofs, string(hash))
	mdr.mutProofs.Unlock()
}

func (mdr *missingDataResolver) allHeadersReceived() bool {
	mdr.mutHeaders.RLock()
	defer mdr.mutHeaders.RUnlock()

	return len(mdr.missingHeaders) == 0
}

func (mdr *missingDataResolver) allProofsReceived() bool {
	mdr.mutProofs.RLock()
	defer mdr.mutProofs.RUnlock()

	return len(mdr.missingProofs) == 0
}

func (mdr *missingDataResolver) allDataReceived() bool {
	return mdr.allHeadersReceived() && mdr.allProofsReceived()
}

func (mdr *missingDataResolver) receivedProof(proof data.HeaderProofHandler) {
	mdr.markProofReceived(proof.GetHeaderHash())
}

func (mdr *missingDataResolver) receivedHeader(_ data.HeaderHandler, headerHash []byte) {
	mdr.markHeaderReceived(headerHash)
}

func (mdr *missingDataResolver) monitorReceivedData() {
	mdr.headersPool.RegisterHandler(mdr.receivedHeader)
	mdr.proofsPool.RegisterHandler(mdr.receivedProof)
}

func (mdr *missingDataResolver) requestHeaderIfNeeded(
	shardID uint32,
	headerHash []byte,
) {
	_, err := mdr.headersPool.GetHeaderByHash(headerHash)
	if err == nil {
		return
	}

	added := mdr.addMissingHeader(headerHash)
	if !added {
		return
	}

	if shardID == core.MetachainShardId {
		go mdr.requestHandler.RequestMetaHeader(headerHash)
	} else {
		go mdr.requestHandler.RequestShardHeader(shardID, headerHash)
	}
}

func (mdr *missingDataResolver) requestProofIfNeeded(shardID uint32, headerHash []byte) {
	if mdr.proofsPool.HasProof(shardID, headerHash) {
		return
	}

	added := mdr.addMissingProof(shardID, headerHash)
	if !added {
		return
	}
	go mdr.requestHandler.RequestEquivalentProofByHash(shardID, headerHash)

}

// todo: maybe use channels instead of polling
func (mdr *missingDataResolver) waitForMissingData(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		if mdr.allDataReceived() {
			return nil
		}

		if time.Now().After(deadline) {
			return errTimeoutWaitingForMissingData
		}
		time.Sleep(checkMissingDataStep)
	}
}
