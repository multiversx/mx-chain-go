package track

import (
	"bytes"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
)

const maxMetaBlocksScannedForInclusion = 16
const maxOwnDescendantsScannedForInclusion = 8

// metaReconciliationEvidenceDepth closes the depth-1 double-extension ambiguity.
const metaReconciliationEvidenceDepth = 2

// ArgsMetaFinalityView holds the pools the meta finality view reads from
type ArgsMetaFinalityView struct {
	HeadersPool dataRetriever.HeadersPool
	ProofsPool  dataRetriever.ProofsPool
}

// metaFinalityView answers meta chain finality questions from the pools alone, so that both the
// processor side and the sync side share one definition; it is stateless and holds no locks
type metaFinalityView struct {
	headersPool dataRetriever.HeadersPool
	proofsPool  dataRetriever.ProofsPool
}

// NewMetaFinalityView creates a view over the node's subjective meta chain finality
func NewMetaFinalityView(args ArgsMetaFinalityView) (*metaFinalityView, error) {
	if check.IfNil(args.HeadersPool) {
		return nil, ErrNilHeadersPool
	}
	if check.IfNil(args.ProofsPool) {
		return nil, ErrNilProofsPool
	}

	return &metaFinalityView{
		headersPool: args.HeadersPool,
		proofsPool:  args.ProofsPool,
	}, nil
}

// IsMetaHeaderHeldFinal reports ordinary finality unless a V3 sibling requires reconciliation.
func (mfv *metaFinalityView) IsMetaHeaderHeldFinal(header data.HeaderHandler, headerHash []byte) bool {
	return isMetaHeaderHeldFinal(mfv.headersPool, mfv.proofsPool, header, headerHash)
}

func isMetaHeaderHeldFinal(
	headersPool dataRetriever.HeadersPool,
	proofsPool dataRetriever.ProofsPool,
	header data.HeaderHandler,
	headerHash []byte,
) bool {
	return isMetaHeaderHeldFinalWithEvidence(
		proofsPool,
		header,
		headerHash,
		func(hash []byte, nonce uint64) data.HeaderHandler {
			return getMetaHeaderFromPool(headersPool, hash, nonce)
		},
		func(nonce uint64, parentHash []byte, depth int) bool {
			return hasProofedMetaDescendants(headersPool, proofsPool, nonce, parentHash, depth)
		},
	)
}

func isMetaHeaderHeldFinalWithEvidence(
	proofsPool dataRetriever.ProofsPool,
	header data.HeaderHandler,
	headerHash []byte,
	getHeader func(hash []byte, nonce uint64) data.HeaderHandler,
	hasDescendants func(nonce uint64, parentHash []byte, depth int) bool,
) bool {
	if check.IfNil(header) || len(headerHash) == 0 {
		return false
	}
	if header.GetShardID() != core.MetachainShardId {
		return false
	}
	if !proofsPool.HasProof(core.MetachainShardId, headerHash) {
		return false
	}

	requiresReconciliation := header.IsHeaderV3() && hasProofedMetaSibling(proofsPool, header, headerHash, getHeader)
	if isMetaHeaderInstantlyFinal(proofsPool, header, requiresReconciliation, getHeader) {
		return true
	}

	if requiresReconciliation {
		return hasDescendants(header.GetNonce()+1, headerHash, metaReconciliationEvidenceDepth)
	}

	return hasDescendants(header.GetNonce()+1, headerHash, 1)
}

// isInstantlyFinal covers the non contended case; the parent needs no recursion of its own, since a
// proofed header is itself the proofed child that settles its parent
func isMetaHeaderInstantlyFinal(
	proofsPool dataRetriever.ProofsPool,
	header data.HeaderHandler,
	requiresReconciliation bool,
	getHeader func(hash []byte, nonce uint64) data.HeaderHandler,
) bool {
	if header.GetNonce() == 0 {
		return false
	}

	parentHash := header.GetPrevHash()
	parent := getHeader(parentHash, header.GetNonce()-1)
	if check.IfNil(parent) {
		return false
	}

	if common.IsContendedHeader(header, parent) {
		return false
	}

	if requiresReconciliation {
		return false
	}

	return proofsPool.HasProof(core.MetachainShardId, parentHash)
}

func hasProofedMetaSibling(
	proofsPool dataRetriever.ProofsPool,
	header data.HeaderHandler,
	headerHash []byte,
	getHeader func(hash []byte, nonce uint64) data.HeaderHandler,
) bool {
	proofs, err := proofsPool.GetProofsByNonce(header.GetNonce(), core.MetachainShardId)
	if err != nil || len(proofs) <= 1 {
		return false
	}

	for _, proof := range proofs {
		if check.IfNil(proof) || proof.GetHeaderNonce() != header.GetNonce() || proof.GetHeaderShardId() != core.MetachainShardId {
			continue
		}

		siblingHash := proof.GetHeaderHash()
		if bytes.Equal(siblingHash, headerHash) {
			continue
		}

		sibling := getHeader(siblingHash, header.GetNonce())
		if check.IfNil(sibling) {
			return true
		}
		if bytes.Equal(sibling.GetPrevHash(), header.GetPrevHash()) {
			return true
		}
	}

	return false
}

func getMetaHeaderFromPool(
	headersPool dataRetriever.HeadersPool,
	headerHash []byte,
	nonce uint64,
) data.HeaderHandler {
	header, err := headersPool.GetHeaderByHash(headerHash)
	if err != nil || check.IfNil(header) || header.GetShardID() != core.MetachainShardId || header.GetNonce() != nonce {
		return nil
	}

	return header
}

// IsIncludedInHeldFinalMetaBlock returns true if a meta block the node holds final references the
// given shard header or one of its descendants on the same branch. Two scan windows: ascending over
// the caller's range (a resumable cursor window, so successive calls cover the whole gap up to the
// pool head) and descending from the pool head.
func (mfv *metaFinalityView) IsIncludedInHeldFinalMetaBlock(shardID uint32, headerHash []byte, nonce uint64, ascendingFrom uint64, ascendingTo uint64) bool {
	if len(headerHash) == 0 {
		return false
	}

	metaNonces := mfv.headersPool.Nonces(core.MetachainShardId)
	if len(metaNonces) == 0 {
		return false
	}

	branch := mfv.ownBranchHashes(shardID, headerHash, nonce)
	highestMetaNonce := highestNonce(metaNonces)
	ascendingEnd := ascendingTo
	// defensive bound: the per call work stays within the window budget whatever the caller sends
	if maxEnd := ascendingFrom + maxMetaBlocksScannedForInclusion - 1; ascendingEnd > maxEnd {
		ascendingEnd = maxEnd
	}

	for metaNonce := ascendingFrom; metaNonce <= ascendingEnd && metaNonce <= highestMetaNonce; metaNonce++ {
		if mfv.holdsFinalMetaBlockReferencing(metaNonce, shardID, branch) {
			return true
		}
	}

	for scanned := uint64(0); scanned < maxMetaBlocksScannedForInclusion; scanned++ {
		metaNonce := highestMetaNonce - scanned
		if metaNonce <= ascendingEnd {
			break
		}

		if mfv.holdsFinalMetaBlockReferencing(metaNonce, shardID, branch) {
			return true
		}
	}

	return false
}

func (mfv *metaFinalityView) holdsFinalMetaBlockReferencing(metaNonce uint64, shardID uint32, branch [][]byte) bool {
	headers, hashes, err := mfv.headersPool.GetHeadersByNonceAndShardId(metaNonce, core.MetachainShardId)
	if err != nil {
		return false
	}

	for i, header := range headers {
		if check.IfNil(header) {
			continue
		}

		metaHeader, ok := header.(data.MetaHeaderHandler)
		if !ok {
			continue
		}

		// in memory check first, so the finality pool reads are paid only for the notarizing block
		if !referencesAnyOf(metaHeader, shardID, branch) {
			continue
		}

		if mfv.IsMetaHeaderHeldFinal(header, hashes[i]) {
			return true
		}
	}

	return false
}

// ownBranchHashes collects the header and its descendants on the same branch; a held final meta
// referencing any of them settles the header, since meta notarizes a descendant only after it
func (mfv *metaFinalityView) ownBranchHashes(shardID uint32, headerHash []byte, nonce uint64) [][]byte {
	branch := [][]byte{headerHash}
	parents := branch

	for depth := uint64(1); depth <= maxOwnDescendantsScannedForInclusion; depth++ {
		children, childrenHashes, err := mfv.headersPool.GetHeadersByNonceAndShardId(nonce+depth, shardID)
		if err != nil {
			break
		}

		descendants := make([][]byte, 0, len(children))
		for i, child := range children {
			if check.IfNil(child) || !containsHash(parents, child.GetPrevHash()) {
				continue
			}

			descendants = append(descendants, childrenHashes[i])
		}

		if len(descendants) == 0 {
			break
		}

		branch = append(branch, descendants...)
		parents = descendants
	}

	return branch
}

// IsDeadMetaBlock returns true if the authority provably built past the meta block on another
// branch: a doubly proofed foreign-parent extension at the next nonce, none of its own
func (mfv *metaFinalityView) IsDeadMetaBlock(headerHash []byte, nonce uint64) bool {
	if len(headerHash) == 0 || nonce == 0 {
		return false
	}

	childNonce := nonce + 1
	children, childrenHashes, err := mfv.headersPool.GetHeadersByNonceAndShardId(childNonce, core.MetachainShardId)
	if err != nil {
		return false
	}

	foreignSettled := false
	for i, child := range children {
		if check.IfNil(child) || bytes.Equal(child.GetPrevHash(), headerHash) {
			continue
		}
		if !mfv.proofsPool.HasProof(core.MetachainShardId, childrenHashes[i]) {
			continue
		}
		if hasProofedMetaDescendants(mfv.headersPool, mfv.proofsPool, childNonce+1, childrenHashes[i], 1) {
			foreignSettled = true
			break
		}
	}
	if !foreignSettled {
		return false
	}

	// a doubly proofed own extension keeps the verdict subjective, the accepted depth-2 residual
	return !HasMetaReconciliationEvidence(mfv.headersPool, mfv.proofsPool, nonce, headerHash)
}

// HasMetaReconciliationEvidence reports whether two proofed descendants extend the header.
func HasMetaReconciliationEvidence(
	headersPool dataRetriever.HeadersPool,
	proofsPool dataRetriever.ProofsPool,
	headerNonce uint64,
	headerHash []byte,
) bool {
	return hasProofedMetaDescendants(
		headersPool,
		proofsPool,
		headerNonce+1,
		headerHash,
		metaReconciliationEvidenceDepth,
	)
}

func hasProofedMetaDescendants(
	headersPool dataRetriever.HeadersPool,
	proofsPool dataRetriever.ProofsPool,
	nonce uint64,
	parentHash []byte,
	depth int,
) bool {
	children, childrenHashes, err := headersPool.GetHeadersByNonceAndShardId(nonce, core.MetachainShardId)
	if err != nil {
		return false
	}

	for i, child := range children {
		if check.IfNil(child) || !bytes.Equal(child.GetPrevHash(), parentHash) {
			continue
		}
		if !proofsPool.HasProof(core.MetachainShardId, childrenHashes[i]) {
			continue
		}
		if depth <= 1 || hasProofedMetaDescendants(headersPool, proofsPool, nonce+1, childrenHashes[i], depth-1) {
			return true
		}
	}

	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (mfv *metaFinalityView) IsInterfaceNil() bool {
	return mfv == nil
}

func referencesAnyOf(metaHeader data.MetaHeaderHandler, shardID uint32, hashes [][]byte) bool {
	for _, shardInfo := range process.GetShardHeadersReferencedByMeta(metaHeader) {
		if shardInfo.GetShardID() != shardID {
			continue
		}

		if containsHash(hashes, shardInfo.GetHeaderHash()) {
			return true
		}
	}

	return false
}

func containsHash(hashes [][]byte, hash []byte) bool {
	for _, current := range hashes {
		if bytes.Equal(current, hash) {
			return true
		}
	}

	return false
}

func highestNonce(nonces []uint64) uint64 {
	highest := nonces[0]
	for _, nonce := range nonces {
		if nonce > highest {
			highest = nonce
		}
	}

	return highest
}
