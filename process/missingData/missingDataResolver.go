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

// MissingDataResolverArgs holds the arguments needed to create a MissingDataResolver
type MissingDataResolverArgs struct {
	HeadersPool        dataRetriever.HeadersPool
	ProofsPool         dataRetriever.ProofsPool
	RequestHandler     process.RequestHandler
	BlockDataRequester process.BlockDataRequester
}

type missingDataResolver struct {
	mutHeaders         sync.RWMutex
	missingHeaders     map[string]struct{}
	mutProofs          sync.RWMutex
	missingProofs      map[string]struct{}
	headersPool        dataRetriever.HeadersPool
	proofsPool         dataRetriever.ProofsPool
	requestHandler     process.RequestHandler
	blockDataRequester process.BlockDataRequester
}

// NewMissingDataResolver creates a new instance of missingDataResolver.
func NewMissingDataResolver(args MissingDataResolverArgs) (*missingDataResolver, error) {
	if check.IfNil(args.HeadersPool) {
		return nil, process.ErrNilHeadersDataPool
	}
	if check.IfNil(args.ProofsPool) {
		return nil, process.ErrNilProofsPool
	}
	if check.IfNil(args.RequestHandler) {
		return nil, process.ErrNilRequestHandler
	}
	if check.IfNil(args.BlockDataRequester) {
		return nil, process.ErrNilBlockDataRequester
	}

	mdt := &missingDataResolver{
		missingHeaders:     make(map[string]struct{}),
		missingProofs:      make(map[string]struct{}),
		headersPool:        args.HeadersPool,
		proofsPool:         args.ProofsPool,
		requestHandler:     args.RequestHandler,
		blockDataRequester: args.BlockDataRequester,
	}

	mdt.monitorReceivedData()
	return mdt, nil
}

// RequestMissingMetaHeadersBlocking requests the missing meta headers and proofs for the given shard header.
func (mdr *missingDataResolver) RequestMissingMetaHeadersBlocking(
	shardHeader data.ShardHeaderHandler,
	timeout time.Duration,
) error {
	err := mdr.RequestMissingMetaHeaders(shardHeader)
	if err != nil {
		return err
	}

	return mdr.WaitForMissingData(timeout)
}

// RequestMissingMetaHeaders requests the missing meta headers and proofs for the given shard header.
func (mdr *missingDataResolver) RequestMissingMetaHeaders(
	shardHeader data.ShardHeaderHandler,
) error {
	if check.IfNil(shardHeader) {
		return process.ErrNilBlockHeader
	}

	metaBlockHashes := shardHeader.GetMetaBlockHashes()
	for i := 0; i < len(metaBlockHashes); i++ {
		mdr.requestHeaderIfNeeded(core.MetachainShardId, metaBlockHashes[i])
		mdr.requestProofIfNeeded(core.MetachainShardId, metaBlockHashes[i])
	}
	return nil
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

// WaitForMissingData waits until all missing data is received or the timeout is reached.
// TODO: maybe use channels instead of polling
func (mdr *missingDataResolver) WaitForMissingData(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	haveTime := func() time.Duration {
		return time.Until(deadline)
	}

	// check if requested miniBlocks and transactions are ready
	err := mdr.blockDataRequester.IsDataPreparedForProcessing(haveTime)
	if err != nil {
		return err
	}

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

// Reset clears the internal state of the missingDataResolver.
func (mdr *missingDataResolver) Reset() {
	mdr.mutHeaders.Lock()
	mdr.missingHeaders = make(map[string]struct{})
	mdr.mutHeaders.Unlock()

	mdr.mutProofs.Lock()
	mdr.missingProofs = make(map[string]struct{})
	mdr.mutProofs.Unlock()

	mdr.blockDataRequester.Reset()
}

// IsInterfaceNil returns true if there is no value under the interface
func (mdr *missingDataResolver) IsInterfaceNil() bool {
	return mdr == nil
}
