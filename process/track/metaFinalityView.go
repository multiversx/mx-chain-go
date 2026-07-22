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

// IsMetaHeaderHeldFinal returns true if the header is proofed and either settled by a proofed child
// or non contended over a proofed parent
func (mfv *metaFinalityView) IsMetaHeaderHeldFinal(header data.HeaderHandler, headerHash []byte) bool {
	if check.IfNil(header) || len(headerHash) == 0 {
		return false
	}
	if header.GetShardID() != core.MetachainShardId {
		return false
	}
	if !mfv.proofsPool.HasProof(core.MetachainShardId, headerHash) {
		return false
	}

	if mfv.isInstantlyFinal(header) {
		return true
	}

	return hasProofedChildInPool(mfv.headersPool, mfv.proofsPool, core.MetachainShardId, headerHash, header.GetNonce()+1)
}

// isInstantlyFinal covers the non contended case; the parent needs no recursion of its own, since a
// proofed header is itself the proofed child that settles its parent
func (mfv *metaFinalityView) isInstantlyFinal(header data.HeaderHandler) bool {
	parentHash := header.GetPrevHash()
	parent, err := mfv.headersPool.GetHeaderByHash(parentHash)
	if err != nil || check.IfNil(parent) || parent.GetNonce()+1 != header.GetNonce() {
		return false
	}

	if common.IsContendedHeader(header, parent) {
		return false
	}

	return mfv.proofsPool.HasProof(core.MetachainShardId, parentHash)
}

// IsIncludedInHeldFinalMetaBlock returns true if a meta block the node holds final references the
// given shard header or one of its descendants on the same branch
func (mfv *metaFinalityView) IsIncludedInHeldFinalMetaBlock(shardID uint32, headerHash []byte, nonce uint64) bool {
	if len(headerHash) == 0 {
		return false
	}

	metaNonces := mfv.headersPool.Nonces(core.MetachainShardId)
	if len(metaNonces) == 0 {
		return false
	}

	branch := mfv.ownBranchHashes(shardID, headerHash, nonce)
	highestMetaNonce := highestNonce(metaNonces)

	for scanned := uint64(0); scanned < maxMetaBlocksScannedForInclusion; scanned++ {
		metaNonce := highestMetaNonce - scanned
		if mfv.holdsFinalMetaBlockReferencing(metaNonce, shardID, branch) {
			return true
		}

		if metaNonce == 0 {
			break
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

// HasHeldFinalCompetitorAtNonce returns true if the node holds a different meta header final at the
// given header's nonce while no longer holding that header final
func (mfv *metaFinalityView) HasHeldFinalCompetitorAtNonce(metaHeader data.HeaderHandler, metaHash []byte) bool {
	if check.IfNil(metaHeader) || len(metaHash) == 0 {
		return false
	}
	if metaHeader.GetShardID() != core.MetachainShardId {
		return false
	}

	proofs, err := mfv.proofsPool.GetProofsByNonce(metaHeader.GetNonce(), core.MetachainShardId)
	if err != nil || len(proofs) <= 1 {
		return false
	}

	// the verdict is exclusive on purpose: while the node still holds its own header final there is
	// nothing to converge away from, even when a sibling looks final as well
	if mfv.IsMetaHeaderHeldFinal(metaHeader, metaHash) {
		return false
	}

	for _, proof := range proofs {
		if check.IfNil(proof) {
			continue
		}

		competitorHash := proof.GetHeaderHash()
		if bytes.Equal(competitorHash, metaHash) {
			continue
		}

		competitor, errGet := mfv.headersPool.GetHeaderByHash(competitorHash)
		if errGet != nil {
			continue
		}

		if mfv.IsMetaHeaderHeldFinal(competitor, competitorHash) {
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
