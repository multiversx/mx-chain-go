package chainSimulator

import (
	"errors"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	mxProcess "github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/asyncExecution/cache"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
)

// errCannotAssembleBlockBody is returned when a node's block processor does not expose the
// pool-based body assembly a follower catch-up needs.
var errCannotAssembleBlockBody = errors.New("block processor cannot assemble a block body from the pool")

// maxTipAndSource returns the highest committed nonce among the shard's consensus nodes and the node
// holding it. Returns (0, nil) when no node has committed a block. Must be called under s.mutex.
func (s *simulator) maxTipAndSource(shardID uint32) (uint64, process.NodeHandler) {
	tip := uint64(0)
	var source process.NodeHandler
	for _, n := range s.consensusNodes[shardID] {
		if check.IfNilReflect(n) {
			continue
		}
		if h := currentNonceOf(n); h > tip {
			tip = h
			source = n
		}
	}

	return tip, source
}

func currentNonceOf(node process.NodeHandler) uint64 {
	header := node.GetChainHandler().GetCurrentBlockHeader()
	if check.IfNil(header) {
		return 0
	}

	return header.GetNonce()
}

// maxTipOf returns the highest committed nonce among the stepped nodes.
func maxTipOf(driverNodes []process.NodeHandler) uint64 {
	tip := uint64(0)
	for _, n := range driverNodes {
		if h := currentNonceOf(n); h > tip {
			tip = h
		}
	}

	return tip
}

// syncBehindNodes commits, on every node the manual consensus drive left behind, the blocks it is
// missing up to the shard tip. The simulator bootstrapper is a synchronized no-op, so this provides
// the follower bootstrap needed before the next round. Must be called under s.mutex after consensus
// stepping has finished.
func (s *simulator) syncBehindNodes(shardID uint32) {
	tip, source := s.maxTipAndSource(shardID)
	if tip == 0 || check.IfNilReflect(source) {
		return
	}

	for _, n := range s.consensusNodes[shardID] {
		if check.IfNilReflect(n) || n == source {
			continue
		}
		// Advance one block at a time; stop at the first nonce this node cannot yet commit (a missing
		// piece that has not propagated) and let the next round retry from there. Also stop if a commit
		// did not advance the head, so a non-advancing commit can never spin this loop under s.mutex.
		for currentNonceOf(n) < tip {
			before := currentNonceOf(n)
			err := s.commitMissingBlock(n, source, shardID, before+1)
			if err != nil {
				log.Debug("syncBehindNodes: could not commit missing block on behind node",
					"shard", shardID, "nonce", before+1, "tip", tip, "error", err)
				break
			}
			if currentNonceOf(n) <= before {
				break
			}
		}
	}

}

// commitMissingBlock applies the shard block at the given nonce to a behind node by committing it
// through the node's own block processor, sourcing the header/proof from the synced tip node when the
// behind node did not receive them. Returns an error (stopping that node's catch-up for this round) if
// any required piece is unavailable. HeaderV3 (Supernova) blocks are replayed through the deferred
// execution flow the consensus followers use (VerifyBlockProposal + AddPairForExecution +
// CommitBlock); older headers go through the legacy ProcessBlock + CommitBlock pair — the legacy
// path rejects V3 headers outright (e.g. checkScheduledData wants the AdditionalData field V3
// headers no longer carry).
func (s *simulator) commitMissingBlock(node, source process.NodeHandler, shardID uint32, nonce uint64) error {
	header, headerHash, err := committedHeaderAtNonce(source, nonce)
	if err != nil {
		return err
	}

	headers := node.GetDataComponents().Datapool().Headers()
	if _, err = headers.GetHeaderByHash(headerHash); err != nil {
		headers.AddHeader(headerHash, header)
	}

	if source.GetCoreComponents().EnableEpochsHandler().IsFlagEnabledInEpoch(common.AndromedaFlag, header.GetEpoch()) {
		proof, proofErr := source.GetDataComponents().Datapool().Proofs().GetProof(shardID, headerHash)
		if proofErr != nil {
			return proofErr
		}

		proofs := node.GetDataComponents().Datapool().Proofs()
		if !proofs.HasProof(shardID, headerHash) {
			proofs.AddProof(proof)
		}
	}

	// The behind node may have missed some of the block's miniblocks (e.g. the epoch-start meta block's
	// peer/reward miniblocks). Copy any it lacks from the synced node so the body assembled from its pool
	// matches the header — otherwise ProcessBlock rejects it with ErrHeaderBodyMismatch.
	ensureMiniBlocks(node, source, header)

	body, err := blockBodyFromPool(node, header)
	if err != nil {
		return err
	}

	processor := node.GetProcessComponents().BlockProcessor()
	haveTime := func() time.Duration { return time.Hour }

	if header.IsHeaderV3() {
		// the deferred-execution replay path, exactly what subroundBlock/subroundEndRound do on a
		// follower: verify the proposal, queue the pair for async execution, commit
		err = processor.VerifyBlockProposal(header, body, haveTime)
		if err != nil {
			return err
		}

		err = node.GetProcessComponents().ExecutionManager().AddPairForExecution(cache.HeaderBodyPair{
			HeaderHash: headerHash,
			Header:     header,
			Body:       body,
		})
		if err != nil {
			return err
		}

		return processor.CommitBlock(header, body)
	}

	// Consensus driving stops once the first group member commits. A lagging member may therefore
	// have processed this proposal but not reached EndRound, leaving its account state dirty. The
	// production chronology reverts that speculative state when the round is extended; catch-up
	// must do the same before replaying the canonical committed block.
	processor.RevertCurrentBlock()

	err = processor.ProcessBlock(header, body, haveTime)
	if err != nil {
		return err
	}

	return processor.CommitBlock(header, body)
}

// committedHeaderAtNonce walks the source's committed chain backwards from its tip. Looking up
// headers only by nonce is ambiguous when failed rounds left competing proposals in the pool; the
// prev-hash chain identifies the block the source actually committed. Committed headers are removed
// from the live pool, so old links fall back to the source node's persistent header storage.
func committedHeaderAtNonce(source process.NodeHandler, nonce uint64) (data.HeaderHandler, []byte, error) {
	chain := source.GetChainHandler()
	header := chain.GetCurrentBlockHeader()
	if check.IfNil(header) || header.GetNonce() < nonce {
		return nil, nil, fmt.Errorf("committed header at nonce %d is unavailable", nonce)
	}
	hash := chain.GetCurrentBlockHeaderHash()

	for header.GetNonce() > nonce {
		hash = header.GetPrevHash()
		var err error
		header, err = committedHeaderByHash(source, hash)
		if err != nil {
			return nil, nil, err
		}
	}

	if header.GetNonce() != nonce {
		return nil, nil, fmt.Errorf("committed header at nonce %d is unavailable", nonce)
	}

	return header, hash, nil
}

func committedHeaderByHash(source process.NodeHandler, hash []byte) (data.HeaderHandler, error) {
	header, poolErr := source.GetDataComponents().Datapool().Headers().GetHeaderByHash(hash)
	if poolErr == nil {
		return header, nil
	}

	header, storageErr := mxProcess.GetHeaderFromStorage(
		source.GetShardCoordinator().SelfId(),
		hash,
		source.GetCoreComponents().InternalMarshalizer(),
		source.GetDataComponents().StorageService(),
	)
	if storageErr != nil {
		return nil, fmt.Errorf(
			"committed header %x is unavailable in pool (%v) and storage: %w",
			hash,
			poolErr,
			storageErr,
		)
	}

	return header, nil
}

// ensureMiniBlocks copies into the node's pool every miniblock the header declares that the node is
// missing. A committed source removes processed miniblocks from its live pool, so use its
// MiniBlocksProvider, which falls back to storage.
func ensureMiniBlocks(node, source process.NodeHandler, header data.HeaderHandler) {
	nodePool := node.GetDataComponents().Datapool().MiniBlocks()
	sourcePool := source.GetDataComponents().Datapool().MiniBlocks()
	sourceProvider := source.GetDataComponents().MiniBlocksProvider()

	for _, mbHeader := range header.GetMiniBlockHeaderHandlers() {
		hash := mbHeader.GetHash()
		if _, ok := nodePool.Get(hash); ok {
			continue
		}

		obj, ok := sourcePool.Get(hash)
		miniBlock, _ := obj.(*block.MiniBlock)
		if !ok || miniBlock == nil {
			miniBlocks, missing := sourceProvider.GetMiniBlocks([][]byte{hash})
			if len(missing) != 0 || len(miniBlocks) != 1 {
				continue
			}
			miniBlock = miniBlocks[0].Miniblock
		}

		// Store a shallow struct copy, not the source's live pointer, so the two nodes' pools never share
		// one mutable object (the field values are identical, so the body hash is unchanged).
		miniBlockCopy := *miniBlock
		nodePool.Put(hash, &miniBlockCopy, miniBlockCopy.Size())
	}
}

// blockBodyFromPool assembles the exact body declared by the header from the node's own miniblock
// pool. This is the proposal body required by HeaderV3's VerifyBlockProposal and is also the normal
// body of a legacy header. Rebuilding it here avoids depending on optional block-processor
// interfaces that can be hidden by simulator-only decorators.
func blockBodyFromPool(node process.NodeHandler, header data.HeaderHandler) (data.BodyHandler, error) {
	return proposalBodyFromPool(node, header)
}

// proposalBodyFromPool rebuilds the exact proposal body of a HeaderV3 from the node's miniblock
// pool: one miniblock per declared miniblock header, in header order. A missing miniblock aborts
// the catch-up for this round (the per-round retry re-sources it via ensureMiniBlocks).
func proposalBodyFromPool(node process.NodeHandler, header data.HeaderHandler) (data.BodyHandler, error) {
	pool := node.GetDataComponents().Datapool().MiniBlocks()
	miniBlockHeaders := header.GetMiniBlockHeaderHandlers()

	miniBlocks := make([]*block.MiniBlock, 0, len(miniBlockHeaders))
	for _, mbHeader := range miniBlockHeaders {
		obj, ok := pool.Get(mbHeader.GetHash())
		if !ok {
			return nil, fmt.Errorf("%w: miniblock %x not in pool", errCannotAssembleBlockBody, mbHeader.GetHash())
		}
		miniBlock, ok := obj.(*block.MiniBlock)
		if !ok {
			return nil, fmt.Errorf("%w: miniblock %x has unexpected type", errCannotAssembleBlockBody, mbHeader.GetHash())
		}
		miniBlocks = append(miniBlocks, miniBlock)
	}

	return &block.Body{MiniBlocks: miniBlocks}, nil
}
