package missingData

import (
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("missingData")

const checkMissingDataStep = 10 * time.Millisecond

// ResolverArgs holds the arguments needed to create a Resolver
type ResolverArgs struct {
	HeadersPool        dataRetriever.HeadersPool
	ProofsPool         dataRetriever.ProofsPool
	RequestHandler     process.RequestHandler
	BlockDataRequester process.BlockDataRequester
}

// Resolver is responsible for requesting and tracking missing headers and proofs.
type Resolver struct {
	mutHeaders         sync.RWMutex
	missingHeaders     map[string]struct{}
	mutProofs          sync.RWMutex
	missingProofs      map[string]struct{}
	headersPool        dataRetriever.HeadersPool
	proofsPool         dataRetriever.ProofsPool
	requestHandler     process.RequestHandler
	blockDataRequester process.BlockDataRequester
}

// NewMissingDataResolver creates a new instance of Resolver.
func NewMissingDataResolver(args ResolverArgs) (*Resolver, error) {
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

	r := &Resolver{
		missingHeaders:     make(map[string]struct{}),
		missingProofs:      make(map[string]struct{}),
		headersPool:        args.HeadersPool,
		proofsPool:         args.ProofsPool,
		requestHandler:     args.RequestHandler,
		blockDataRequester: args.BlockDataRequester,
	}

	r.monitorReceivedData()
	return r, nil
}

// RequestMissingMetaHeadersBlocking requests the missing meta headers and proofs for the given shard header.
func (r *Resolver) RequestMissingMetaHeadersBlocking(
	shardHeader data.ShardHeaderHandler,
	timeout time.Duration,
) error {
	err := r.RequestMissingMetaHeaders(shardHeader)
	if err != nil {
		return err
	}

	return r.WaitForMissingData(timeout)
}

// RequestMissingMetaHeaders requests the missing meta headers and proofs for the given shard header.
func (r *Resolver) RequestMissingMetaHeaders(
	shardHeader data.ShardHeaderHandler,
) error {
	if check.IfNil(shardHeader) {
		return process.ErrNilBlockHeader
	}

	metaBlockHashes := shardHeader.GetMetaBlockHashes()
	if shardHeader.IsStartOfEpochBlock() {
		epochStartMetaHash := shardHeader.GetEpochStartMetaHash()
		metaBlockHashes = append(metaBlockHashes, epochStartMetaHash)
	}

	for i := 0; i < len(metaBlockHashes); i++ {
		r.requestHeaderIfNeeded(core.MetachainShardId, metaBlockHashes[i])
		r.requestProofIfNeeded(core.MetachainShardId, metaBlockHashes[i])
	}
	return nil
}

func (r *Resolver) addMissingHeader(hash []byte) bool {
	r.mutHeaders.Lock()
	r.missingHeaders[string(hash)] = struct{}{}
	r.mutHeaders.Unlock()

	// avoid missing notifications if the header just arrived
	_, err := r.headersPool.GetHeaderByHash(hash)
	if err == nil {
		r.mutHeaders.Lock()
		delete(r.missingHeaders, string(hash))
		r.mutHeaders.Unlock()
	}

	return err != nil
}

func (r *Resolver) addMissingProof(shardID uint32, hash []byte) bool {
	r.mutProofs.Lock()
	r.missingProofs[string(hash)] = struct{}{}
	r.mutProofs.Unlock()

	// avoid missing notifications if the proof just arrived
	hasProof := r.proofsPool.HasProof(shardID, hash)
	if hasProof {
		r.mutProofs.Lock()
		delete(r.missingProofs, string(hash))
		r.mutProofs.Unlock()
	}

	return !hasProof
}

func (r *Resolver) markHeaderReceived(hash []byte) {
	r.mutHeaders.Lock()
	delete(r.missingHeaders, string(hash))
	r.mutHeaders.Unlock()
}

func (r *Resolver) markProofReceived(hash []byte) {
	r.mutProofs.Lock()
	delete(r.missingProofs, string(hash))
	r.mutProofs.Unlock()
}

func (r *Resolver) allHeadersReceived() bool {
	r.mutHeaders.RLock()
	defer r.mutHeaders.RUnlock()

	return len(r.missingHeaders) == 0
}

func (r *Resolver) allProofsReceived() bool {
	r.mutProofs.RLock()
	defer r.mutProofs.RUnlock()

	return len(r.missingProofs) == 0
}

func (r *Resolver) allDataReceived() bool {
	return r.allHeadersReceived() && r.allProofsReceived()
}

func (r *Resolver) receivedProof(proof data.HeaderProofHandler) {
	r.markProofReceived(proof.GetHeaderHash())
}

func (r *Resolver) receivedHeader(_ data.HeaderHandler, headerHash []byte) {
	r.markHeaderReceived(headerHash)
}

func (r *Resolver) monitorReceivedData() {
	r.headersPool.RegisterHandler(r.receivedHeader)
	r.proofsPool.RegisterHandler(r.receivedProof)
}

func (r *Resolver) requestHeaderIfNeeded(
	shardID uint32,
	headerHash []byte,
) {
	_, err := r.headersPool.GetHeaderByHash(headerHash)
	if err == nil {
		return
	}

	added := r.addMissingHeader(headerHash)
	if !added {
		return
	}

	if shardID == core.MetachainShardId {
		go r.requestHandler.RequestMetaHeader(headerHash)
	} else {
		go r.requestHandler.RequestShardHeader(shardID, headerHash)
	}
}

func (r *Resolver) requestProofIfNeeded(shardID uint32, headerHash []byte) {
	if r.proofsPool.HasProof(shardID, headerHash) {
		return
	}

	added := r.addMissingProof(shardID, headerHash)
	if !added {
		return
	}

	go r.requestHandler.RequestEquivalentProofByHash(shardID, headerHash)
}

// WaitForMissingData waits until all missing data is received or the timeout is reached.
// TODO: maybe use channels instead of polling
func (r *Resolver) WaitForMissingData(timeout time.Duration) error {
	waitDeadline := time.Now().Add(timeout)

	stepHaveTime := func(stepTimeout time.Duration) func() time.Duration {
		stepDeadline := time.Now().Add(stepTimeout)
		haveTime := func() time.Duration {
			return time.Until(stepDeadline)
		}
		return haveTime
	}

	for {
		err := r.blockDataRequester.IsDataPreparedForProcessing(stepHaveTime(checkMissingDataStep))
		if r.allDataReceived() && err == nil {
			return nil
		}

		if time.Now().After(waitDeadline) {
			return process.ErrTimeIsOut
		}
	}
}

// RequestMissingShardHeadersBlocking requests the missing shard headers and proofs for the given meta header and waits for results.
func (r *Resolver) RequestMissingShardHeadersBlocking(
	metaHeader data.MetaHeaderHandler,
	timeout time.Duration,
) error {
	err := r.RequestMissingShardHeaders(metaHeader)
	if err != nil {
		return err
	}

	return r.WaitForMissingData(timeout)
}

// RequestMissingShardHeaders requests the missing shard headers and proofs for the given meta header.
func (r *Resolver) RequestMissingShardHeaders(
	metaHeader data.MetaHeaderHandler,
) error {
	if check.IfNil(metaHeader) {
		return process.ErrNilMetaBlockHeader
	}

	if metaHeader.IsStartOfEpochBlock() && metaHeader.GetEpochStartHandler() != nil {
		r.requestEpochStartLastFinalizedHeaders(metaHeader.GetEpochStartHandler())
	}

	shardDataProposedNonces := make(map[uint32]uint64)
	for _, shardProposalData := range metaHeader.GetShardInfoProposalHandlers() {
		shardDataProposedNonces[shardProposalData.GetShardID()] = shardProposalData.GetNonce()

		r.requestHeaderIfNeeded(shardProposalData.GetShardID(), shardProposalData.GetHeaderHash())
		r.requestProofIfNeeded(shardProposalData.GetShardID(), shardProposalData.GetHeaderHash())
	}

	shardDataFinalizedNonces := getShardDataFinalizedNonces(metaHeader.GetShardInfoHandlers())
	r.requestNonceGapsIfNeeded(shardDataFinalizedNonces, shardDataProposedNonces)

	return nil
}

func (r *Resolver) requestEpochStartLastFinalizedHeaders(epochStartHandler data.EpochStartHandler) {
	for _, finalizedHdr := range epochStartHandler.GetLastFinalizedHeaderHandlers() {
		r.requestHeaderIfNeeded(finalizedHdr.GetShardID(), finalizedHdr.GetHeaderHash())
		r.requestProofIfNeeded(finalizedHdr.GetShardID(), finalizedHdr.GetHeaderHash())
	}
}

func getShardDataFinalizedNonces(shardInfoHandlers []data.ShardDataHandler) map[uint32]uint64 {
	shardDataFinalizedNonces := make(map[uint32]uint64)
	for _, shardData := range shardInfoHandlers {
		shardDataFinalizedNonces[shardData.GetShardID()] = shardData.GetNonce()
	}
	return shardDataFinalizedNonces
}

func (r *Resolver) requestNonceGapsIfNeeded(shardDataFinalizedNonces, shardDataProposedNonces map[uint32]uint64) {
	for shardID, proposedNonce := range shardDataProposedNonces {
		lastFinalizedNonce, found := shardDataFinalizedNonces[shardID]
		if !found {
			log.Warn("Resolver.requestNonceGapsIfNeeded: shard found in shardDataFinalizedNonces", "shard", shardID)
			continue
		}

		nonceGaps := proposedNonce - lastFinalizedNonce
		if nonceGaps < 2 {
			continue
		}

		r.requestShardHeadersAndProofsByNonce(shardID, lastFinalizedNonce+1, proposedNonce)
	}
}

// requestShardHeadersAndProofsByNonce will request shard headers and proofs if needed without blocking
func (r *Resolver) requestShardHeadersAndProofsByNonce(shardID uint32, startNonce, endNonce uint64) {
	for shardNonceToRequest := startNonce; shardNonceToRequest < endNonce; shardNonceToRequest++ {
		r.requestShardHeaderByNonceIfNeeded(shardID, shardNonceToRequest)
		r.requestShardProofByNonceIfNeeded(shardID, shardNonceToRequest)
	}
}

func (r *Resolver) requestShardHeaderByNonceIfNeeded(shardID uint32, nonce uint64) {
	if _, _, err := r.headersPool.GetHeadersByNonceAndShardId(nonce, shardID); err == nil {
		return
	}

	go r.requestHandler.RequestShardHeaderByNonce(shardID, nonce)
}

func (r *Resolver) requestShardProofByNonceIfNeeded(shardID uint32, nonce uint64) {
	if _, err := r.proofsPool.GetProofByNonce(nonce, shardID); err == nil {
		return
	}

	go r.requestHandler.RequestEquivalentProofByNonce(shardID, nonce)
}

// RequestBlockTransactions requests the transactions for the given block body.
func (r *Resolver) RequestBlockTransactions(body *block.Body) {
	r.blockDataRequester.RequestBlockTransactions(body)
}

// RequestMiniBlocksAndTransactions requests mini blocks and transactions if missing
func (r *Resolver) RequestMiniBlocksAndTransactions(header data.HeaderHandler) {
	r.blockDataRequester.RequestMiniBlocksAndTransactions(header)
}

// GetFinalCrossMiniBlockInfoAndRequestMissing returns the final cross mini block infos and requests missing mini blocks and transactions
func (r *Resolver) GetFinalCrossMiniBlockInfoAndRequestMissing(header data.HeaderHandler) []*data.MiniBlockInfo {
	return r.blockDataRequester.GetFinalCrossMiniBlockInfoAndRequestMissing(header)
}

// Reset clears the internal state of the Resolver.
func (r *Resolver) Reset() {
	r.mutHeaders.Lock()
	r.missingHeaders = make(map[string]struct{})
	r.mutHeaders.Unlock()

	r.mutProofs.Lock()
	r.missingProofs = make(map[string]struct{})
	r.mutProofs.Unlock()

	r.blockDataRequester.Reset()
}

// IsInterfaceNil returns true if there is no value under the interface
func (r *Resolver) IsInterfaceNil() bool {
	return r == nil
}
