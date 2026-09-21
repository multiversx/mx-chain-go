package storageBootstrap

import (
	"bytes"
	"errors"
	"fmt"
	"math"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/process/sync/storageBootstrap/metricsLoader"
	"github.com/multiversx/mx-chain-go/storage"
)

var ErrRecoveryCheckpointUnavailable = errors.New("recovery checkpoint unavailable")

const maxRecoverySelectorSpan = 1_000_000
const maxRecoverySelectorCopies = 64

type recoverySelectorBounds struct {
	maxOwnNonce uint64
	maxCross    map[uint32]uint64
}

func (st *storageBootstrapper) checkRecoveryHeader(header data.HeaderHandler, hash []byte) error {
	if check.IfNil(header) {
		return process.ErrNilHeaderHandler
	}
	if st.recoveryCheckpoint.IsDiscarded(header.GetRound(), header.GetShardID(), hash) {
		return common.ErrRoundExcluded
	}
	return nil
}

func (st *storageBootstrapper) shouldRecoverCheckpoint() (bool, error) {
	highestRound := st.bootStorer.GetHighestRound()
	if highestRound < int64(st.recoveryCheckpoint.Round) {
		return false, fmt.Errorf("%w: bootstrap round %d is before target %d", ErrRecoveryCheckpointUnavailable, highestRound, st.recoveryCheckpoint.Round)
	}
	return uint64(highestRound) <= st.recoveryCheckpoint.ExcludedEnd, nil
}

func (st *storageBootstrapper) verifyRecoveryHeaderHash(header data.HeaderHandler, hash []byte) error {
	if check.IfNil(st.hasher) {
		return ErrRecoveryCheckpointUnavailable
	}
	computed, err := core.CalculateHash(st.marshalizer, st.hasher, header)
	if err != nil || !bytes.Equal(computed, hash) {
		return fmt.Errorf("%w: header hash mismatch: %v", ErrRecoveryCheckpointUnavailable, err)
	}
	return nil
}

func (st *storageBootstrapper) getRecoveryHeader() (bootstrapStorage.BootstrapData, data.HeaderHandler, error) {
	target, err := st.bootStorer.Get(int64(st.recoveryCheckpoint.Round))
	if err != nil {
		return bootstrapStorage.BootstrapData{}, nil, fmt.Errorf("%w: target bootstrap record: %w", ErrRecoveryCheckpointUnavailable, err)
	}
	expectedHash, ok := st.recoveryCheckpoint.HeaderHash(st.shardCoordinator.SelfId())
	if !ok || !bytes.Equal(target.LastHeader.Hash, expectedHash) || target.LastHeader.ShardId != st.shardCoordinator.SelfId() {
		return bootstrapStorage.BootstrapData{}, nil, fmt.Errorf("%w: target bootstrap hash or shard mismatch", ErrRecoveryCheckpointUnavailable)
	}
	header, err := st.bootstrapper.getHeader(expectedHash)
	if err != nil {
		return bootstrapStorage.BootstrapData{}, nil, fmt.Errorf("%w: target header: %w", ErrRecoveryCheckpointUnavailable, err)
	}
	if check.IfNil(header) || !header.IsHeaderV3() || header.GetRound() != st.recoveryCheckpoint.Round ||
		header.GetShardID() != target.LastHeader.ShardId || header.GetNonce() != target.LastHeader.Nonce ||
		header.GetEpoch() != target.LastHeader.Epoch || string(header.GetChainID()) != st.chainID {
		return bootstrapStorage.BootstrapData{}, nil, fmt.Errorf("%w: target header fields mismatch", ErrRecoveryCheckpointUnavailable)
	}
	if err = st.verifyRecoveryHeaderHash(header, expectedHash); err != nil {
		return bootstrapStorage.BootstrapData{}, nil, err
	}
	rootHash, err := st.getRootHashForBlock(header, expectedHash)
	if err != nil || len(rootHash) != 32 {
		return bootstrapStorage.BootstrapData{}, nil, fmt.Errorf("%w: target state root: %v", ErrRecoveryCheckpointUnavailable, err)
	}
	if header.GetShardID() == core.MetachainShardId {
		lastResult, ok := header.GetLastExecutionResultHandler().(data.LastMetaExecutionResultHandler)
		if !ok || check.IfNil(lastResult) || check.IfNil(lastResult.GetExecutionResultHandler()) ||
			len(lastResult.GetExecutionResultHandler().GetValidatorStatsRootHash()) != 32 {
			return bootstrapStorage.BootstrapData{}, nil, fmt.Errorf("%w: target peer state root", ErrRecoveryCheckpointUnavailable)
		}
	}
	if target.LastRound <= 0 || uint64(target.LastRound) >= st.recoveryCheckpoint.Round {
		return bootstrapStorage.BootstrapData{}, nil, fmt.Errorf("%w: target parent bootstrap record is invalid", ErrRecoveryCheckpointUnavailable)
	}
	parent, err := st.bootstrapper.getHeader(header.GetPrevHash())
	if err != nil || check.IfNil(parent) || parent.GetNonce() >= header.GetNonce() || header.GetNonce()-parent.GetNonce() != 1 ||
		parent.GetRound() != uint64(target.LastRound) || parent.GetShardID() != header.GetShardID() ||
		string(parent.GetChainID()) != st.chainID {
		return bootstrapStorage.BootstrapData{}, nil, fmt.Errorf("%w: target parent header: %v", ErrRecoveryCheckpointUnavailable, err)
	}
	if err = st.verifyRecoveryHeaderHash(parent, header.GetPrevHash()); err != nil {
		return bootstrapStorage.BootstrapData{}, nil, err
	}
	proof, err := st.getProofForHeader(expectedHash, header)
	if err != nil || proof == nil || !bytes.Equal(proof.GetHeaderHash(), expectedHash) ||
		proof.GetHeaderRound() != header.GetRound() || proof.GetHeaderNonce() != header.GetNonce() ||
		proof.GetHeaderShardId() != header.GetShardID() || proof.GetHeaderEpoch() != header.GetEpoch() {
		return bootstrapStorage.BootstrapData{}, nil, fmt.Errorf("%w: target proof: %v", ErrRecoveryCheckpointUnavailable, err)
	}
	return target, header, nil
}

func (st *storageBootstrapper) getRecoveryCheckpointState() ([]byte, []byte, uint32, error) {
	_, header, err := st.getRecoveryHeader()
	if err != nil {
		return nil, nil, 0, err
	}
	userRoot, err := st.getRootHashForBlock(header, nil)
	if err != nil {
		return nil, nil, 0, err
	}
	if header.GetShardID() != core.MetachainShardId {
		return userRoot, nil, header.GetEpoch(), nil
	}
	lastResult, ok := header.GetLastExecutionResultHandler().(data.LastMetaExecutionResultHandler)
	if !ok || check.IfNil(lastResult) || check.IfNil(lastResult.GetExecutionResultHandler()) {
		return nil, nil, 0, ErrRecoveryCheckpointUnavailable
	}
	return userRoot, lastResult.GetExecutionResultHandler().GetValidatorStatsRootHash(), header.GetEpoch(), nil
}

func (st *storageBootstrapper) RecoveryCheckpointState() ([]byte, []byte, uint32, error) {
	if st.recoveryCheckpoint == nil {
		return nil, nil, 0, ErrRecoveryCheckpointUnavailable
	}
	return st.getRecoveryCheckpointState()
}

func (st *storageBootstrapper) RecoveryCheckpointRequired() (bool, error) {
	if st.recoveryCheckpoint == nil {
		return false, nil
	}
	return st.shouldRecoverCheckpoint()
}

func (st *storageBootstrapper) collectRecoverySuffix() (recoverySelectorBounds, error) {
	round := st.bootStorer.GetHighestRound()
	bounds := recoverySelectorBounds{maxCross: make(map[uint32]uint64)}
	for uint64(round) > st.recoveryCheckpoint.Round {
		dataAtRound, err := st.bootStorer.Get(round)
		if err != nil {
			return bounds, fmt.Errorf("%w: discarded bootstrap round %d: %w", ErrRecoveryCheckpointUnavailable, round, err)
		}
		if dataAtRound.LastRound >= round || dataAtRound.LastRound < int64(st.recoveryCheckpoint.Round) {
			return bounds, fmt.Errorf("%w: discarded bootstrap ancestry at round %d", ErrRecoveryCheckpointUnavailable, round)
		}
		header, err := st.bootstrapper.getHeader(dataAtRound.LastHeader.Hash)
		if err != nil || check.IfNil(header) || header.GetRound() != uint64(round) ||
			header.GetShardID() != dataAtRound.LastHeader.ShardId || header.GetNonce() != dataAtRound.LastHeader.Nonce {
			return bounds, fmt.Errorf("%w: discarded header at round %d: %v", ErrRecoveryCheckpointUnavailable, round, err)
		}
		if dataAtRound.LastHeader.Nonce > bounds.maxOwnNonce {
			bounds.maxOwnNonce = dataAtRound.LastHeader.Nonce
		}
		for _, info := range dataAtRound.LastCrossNotarizedHeaders {
			if info.Nonce > bounds.maxCross[info.ShardId] {
				bounds.maxCross[info.ShardId] = info.Nonce
			}
		}
		round = dataAtRound.LastRound
	}
	if uint64(round) != st.recoveryCheckpoint.Round {
		return bounds, ErrRecoveryCheckpointUnavailable
	}
	return bounds, nil
}

func (st *storageBootstrapper) removeNonceSelectors(unit storage.Storer, firstNonce uint64, lastNonce uint64) error {
	if lastNonce < firstNonce {
		return nil
	}
	if lastNonce == math.MaxUint64 {
		return ErrRecoveryCheckpointUnavailable
	}
	if lastNonce-firstNonce > maxRecoverySelectorSpan {
		return ErrRecoveryCheckpointUnavailable
	}
	for nonce := firstNonce; nonce <= lastNonce; nonce++ {
		key := st.uint64Converter.ToByteSlice(nonce)
		for copyIndex := 0; copyIndex < maxRecoverySelectorCopies; copyIndex++ {
			err := unit.Has(key)
			if errors.Is(err, storage.ErrKeyNotFound) {
				break
			}
			if err != nil {
				return err
			}
			if err = unit.Remove(key); err != nil {
				return err
			}
			if copyIndex == maxRecoverySelectorCopies-1 {
				return fmt.Errorf("%w: nonce selector %d remains", ErrRecoveryCheckpointUnavailable, nonce)
			}
		}
	}
	return nil
}

func (st *storageBootstrapper) ensureNonceSelector(unit storage.Storer, nonce uint64, hash []byte) error {
	key := st.uint64Converter.ToByteSlice(nonce)
	err := unit.Has(key)
	if err != nil && !errors.Is(err, storage.ErrKeyNotFound) {
		return err
	}
	if err == nil {
		storedHash, getErr := unit.Get(key)
		if getErr != nil {
			return getErr
		}
		if bytes.Equal(storedHash, hash) {
			return nil
		}
	}
	if err = unit.Put(key, hash); err != nil {
		return err
	}
	storedHash, err := unit.Get(key)
	if err != nil || !bytes.Equal(storedHash, hash) {
		return ErrRecoveryCheckpointUnavailable
	}
	return nil
}

func (st *storageBootstrapper) removeDiscardedSelectors(target bootstrapStorage.BootstrapData, bounds recoverySelectorBounds) error {
	maxOwnNonce := max(target.LastHeader.Nonce, bounds.maxOwnNonce)
	if maxOwnNonce == ^uint64(0) || target.LastHeader.Nonce == ^uint64(0) {
		return ErrRecoveryCheckpointUnavailable
	}
	if err := st.removeNonceSelectors(st.headerNonceHashStore, target.LastHeader.Nonce+1, maxOwnNonce+1); err != nil {
		return err
	}
	if err := st.ensureNonceSelector(st.headerNonceHashStore, target.LastHeader.Nonce, target.LastHeader.Hash); err != nil {
		return err
	}

	baseline := make(map[uint32]uint64, len(target.LastCrossNotarizedHeaders))
	for _, info := range target.LastCrossNotarizedHeaders {
		if _, exists := baseline[info.ShardId]; exists {
			return ErrRecoveryCheckpointUnavailable
		}
		baseline[info.ShardId] = info.Nonce
	}
	for shardID, highNonce := range bounds.maxCross {
		lowNonce, ok := baseline[shardID]
		if !ok {
			return ErrRecoveryCheckpointUnavailable
		}
		if highNonce <= lowNonce {
			continue
		}
		unit, err := st.store.GetStorer(dataRetriever.GetHdrNonceHashDataUnit(shardID))
		if err != nil {
			return err
		}
		if err = st.removeNonceSelectors(unit, lowNonce+1, highNonce); err != nil {
			return err
		}
	}
	for _, info := range target.LastCrossNotarizedHeaders {
		unit, err := st.store.GetStorer(dataRetriever.GetHdrNonceHashDataUnit(info.ShardId))
		if err != nil {
			return err
		}
		if err = st.ensureNonceSelector(unit, info.Nonce, info.Hash); err != nil {
			return err
		}
	}
	return nil
}

func (st *storageBootstrapper) loadRecoveryCheckpoint() error {
	target, _, err := st.getRecoveryHeader()
	if err != nil {
		return err
	}
	bounds, err := st.collectRecoverySuffix()
	if err != nil {
		return err
	}
	bootInfos, err := st.getBootInfos(target)
	if err != nil {
		return fmt.Errorf("%w: target ancestry: %w", ErrRecoveryCheckpointUnavailable, err)
	}
	_, numHeaders := metricsLoader.UpdateMetricsFromStorage(st.store, st.uint64Converter, st.marshalizer, st.appStatusHandler, target.LastHeader.Nonce)
	st.blkExecutor.SetNumProcessedObj(numHeaders)
	st.blockTracker.RestoreToGenesis()
	st.forkDetector.RestoreToGenesis()
	if err = st.applyHeaderInfo(target); err != nil {
		return fmt.Errorf("%w: target state: %w", ErrRecoveryCheckpointUnavailable, err)
	}
	if err = st.applyBootInfos(bootInfos); err != nil {
		return fmt.Errorf("%w: target bootstrap state: %w", ErrRecoveryCheckpointUnavailable, err)
	}
	currentHeader, currentHash := st.blkc.GetCurrentBlockHeaderAndHash()
	if check.IfNil(currentHeader) || !bytes.Equal(currentHash, target.LastHeader.Hash) {
		return fmt.Errorf("%w: restored tip mismatch", ErrRecoveryCheckpointUnavailable)
	}
	if err = st.executionManager.RewindExecutionStateToTip(currentHeader, currentHash); err != nil {
		return fmt.Errorf("%w: execution state: %w", ErrRecoveryCheckpointUnavailable, err)
	}
	st.forkDetector.SetFinalToLastCheckpoint()
	finalNonce, finalHash := st.forkDetector.GetHighestFinalBlockNonce(), st.forkDetector.GetHighestFinalBlockHash()
	settledNonce, settledHash := st.forkDetector.GetHighestSettledBlockInfo()
	if finalNonce != target.LastHeader.Nonce || !bytes.Equal(finalHash, target.LastHeader.Hash) ||
		settledNonce != target.LastHeader.Nonce || !bytes.Equal(settledHash, target.LastHeader.Hash) {
		return fmt.Errorf("%w: checkpoint finality mismatch", ErrRecoveryCheckpointUnavailable)
	}
	st.bootstrapper.applyNumPendingMiniBlocks(target.PendingMiniBlocks)
	st.processedMiniBlocksTracker.ConvertSliceToProcessedMiniBlocksMap(target.ProcessedMiniBlocks)
	st.scheduledTxsExecutionHandler.SetScheduledInfo(&process.ScheduledInfo{
		IntermediateTxs: make(map[block.Type][]data.TransactionHandler),
		GasAndFees:      process.GetZeroGasAndFees(),
		MiniBlocks:      make(block.MiniBlockSlice, 0),
	})
	if err = st.removeDiscardedSelectors(target, bounds); err != nil {
		return fmt.Errorf("%w: selector cleanup: %w", ErrRecoveryCheckpointUnavailable, err)
	}
	if err = st.bootStorer.SaveLastRound(int64(st.recoveryCheckpoint.Round)); err != nil {
		return fmt.Errorf("%w: save target round: %w", ErrRecoveryCheckpointUnavailable, err)
	}
	st.highestNonce = target.LastHeader.Nonce
	st.epochNotifier.CheckEpoch(st.blkc.GetCurrentBlockHeader())
	return nil
}
