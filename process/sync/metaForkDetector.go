package sync

import (
	"math"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/process"
)

var _ process.ForkDetector = (*metaForkDetector)(nil)

// metaForkDetector implements the meta fork detector mechanism
type metaForkDetector struct {
	*baseForkDetector
}

// NewMetaForkDetector method creates a new metaForkDetector object
func NewMetaForkDetector(
	roundHandler consensus.RoundHandler,
	blackListHandler process.TimeCacher,
	blockTracker process.BlockTracker,
	genesisTime int64,
	supernovaGenesisTime int64,
	enableEpochsHandler common.EnableEpochsHandler,
	enableRoundsHandler common.EnableRoundsHandler,
	proofsPool process.ProofsPool,
	chainParametersHandler common.ChainParametersHandler,
	processConfigsHandler common.ProcessConfigsHandler,
) (*metaForkDetector, error) {
	if check.IfNil(roundHandler) {
		return nil, process.ErrNilRoundHandler
	}
	if check.IfNil(blackListHandler) {
		return nil, process.ErrNilBlackListCacher
	}
	if check.IfNil(blockTracker) {
		return nil, process.ErrNilBlockTracker
	}
	if check.IfNil(enableEpochsHandler) {
		return nil, process.ErrNilEnableEpochsHandler
	}
	if check.IfNil(enableRoundsHandler) {
		return nil, process.ErrNilEnableRoundsHandler
	}
	if check.IfNil(proofsPool) {
		return nil, process.ErrNilProofsPool
	}
	if check.IfNil(chainParametersHandler) {
		return nil, process.ErrNilChainParametersHandler
	}
	if check.IfNil(processConfigsHandler) {
		return nil, process.ErrNilProcessConfigsHandler
	}

	genesisHdr, _, err := blockTracker.GetSelfNotarizedHeader(core.MetachainShardId, 0)
	if err != nil {
		return nil, err
	}

	bfd := &baseForkDetector{
		roundHandler:           roundHandler,
		blackListHandler:       blackListHandler,
		genesisTime:            genesisTime,
		supernovaGenesisTime:   supernovaGenesisTime,
		blockTracker:           blockTracker,
		genesisNonce:           genesisHdr.GetNonce(),
		genesisRound:           genesisHdr.GetRound(),
		genesisEpoch:           genesisHdr.GetEpoch(),
		enableEpochsHandler:    enableEpochsHandler,
		enableRoundsHandler:    enableRoundsHandler,
		proofsPool:             proofsPool,
		chainParametersHandler: chainParametersHandler,
		processConfigsHandler:  processConfigsHandler,
		shardID:                core.MetachainShardId,
	}

	bfd.headers = make(map[uint64][]*headerInfo)
	bfd.fork.checkpoint = make([]*checkpointInfo, 0)
	checkpoint := &checkpointInfo{
		nonce: bfd.genesisNonce,
		round: bfd.genesisRound,
	}
	bfd.setFinalAndSettledCheckpoint(checkpoint)
	bfd.addCheckpoint(checkpoint)
	bfd.fork.rollBackNonce = math.MaxUint64
	bfd.fork.probableHighestNonce = bfd.genesisNonce
	bfd.fork.highestNonceReceived = bfd.genesisNonce

	mfd := metaForkDetector{
		baseForkDetector: bfd,
	}

	bfd.forkDetector = &mfd

	return &mfd, nil
}

// AddHeader method adds a new header to headers map
func (mfd *metaForkDetector) AddHeader(
	header data.HeaderHandler,
	headerHash []byte,
	state process.BlockHeaderState,
	selfNotarizedHeaders []data.HeaderHandler,
	selfNotarizedHeadersHashes [][]byte,
) error {
	return mfd.addHeader(
		header,
		headerHash,
		state,
		selfNotarizedHeaders,
		selfNotarizedHeadersHashes,
		mfd.doJobOnBHProcessed,
	)
}

func (mfd *metaForkDetector) doJobOnBHProcessed(
	header data.HeaderHandler,
	headerHash []byte,
	_ []data.HeaderHandler,
	_ [][]byte,
) {
	mfd.mutFinalityUpdate.Lock()
	defer mfd.mutFinalityUpdate.Unlock()

	lastCheckpoint := mfd.lastCheckpoint()
	// under Supernova the committed header settles only the block it extends (settle-on-child)
	canSettleLastCheckpoint := !mfd.isSupernovaForHeader(header) || isParentCheckpoint(lastCheckpoint, header)
	if canSettleLastCheckpoint {
		mfd.advanceFinalAndSettledCheckpoint(lastCheckpoint)
	}
	newCheckpoint := &checkpointInfo{nonce: header.GetNonce(), round: header.GetRound(), hash: headerHash}
	mfd.addCheckpoint(newCheckpoint)
	if common.IsProofsFlagEnabledForHeader(mfd.enableEpochsHandler, header) {
		mfd.setInstantFinalCheckpoint(header, headerHash, newCheckpoint)
	}
	mfd.removePastOrInvalidRecords()
	mfd.logFinalityLag()
}

func (mfd *metaForkDetector) computeFinalCheckpoint() {
}

// ReconcileFinalCheckpoint serializes reconciliation with metachain finality updates.
func (mfd *metaForkDetector) ReconcileFinalCheckpoint(nonce uint64) bool {
	mfd.mutFinalityUpdate.Lock()
	defer mfd.mutFinalityUpdate.Unlock()

	return mfd.baseForkDetector.ReconcileFinalCheckpoint(nonce)
}

// ReconcileFinalCheckpointBelow serializes suffix reconciliation with metachain finality updates.
func (mfd *metaForkDetector) ReconcileFinalCheckpointBelow(nonce uint64) bool {
	mfd.mutFinalityUpdate.Lock()
	reconciled, loweredFinal := mfd.reconcileFinalCheckpointRecordsBelow(nonce)
	mfd.mutFinalityUpdate.Unlock()
	if reconciled {
		mfd.finishFinalCheckpointReconciliation(nonce, loweredFinal)
	}

	return reconciled
}

// ReconcileFinalCheckpointFromAuthority serializes authority reconciliation with metachain finality updates.
func (mfd *metaForkDetector) ReconcileFinalCheckpointFromAuthority(nonce uint64, selectedHash []byte) bool {
	if len(selectedHash) == 0 {
		return false
	}

	mfd.mutFinalityUpdate.Lock()
	reconciled, loweredFinal := mfd.reconcileFinalCheckpointFromAuthorityRecords(nonce, selectedHash)
	mfd.mutFinalityUpdate.Unlock()
	if reconciled {
		mfd.finishFinalCheckpointReconciliation(nonce, loweredFinal)
	}

	return reconciled
}

// IsInterfaceNil returns true if there is no value under the interface
func (mfd *metaForkDetector) IsInterfaceNil() bool {
	return mfd == nil
}
