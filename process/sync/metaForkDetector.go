package sync

import (
	"math"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/process"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var _ process.ForkDetector = (*metaForkDetector)(nil)

// metaForkDetector implements the meta fork detector mechanism
type metaForkDetector struct {
	*baseForkDetector
}

// NewMetaForkDetector method creates a new metaForkDetector object
func NewMetaForkDetector(
	log logger.Logger,
	roundHandler consensus.RoundHandler,
	blackListHandler process.TimeCacher,
	blockTracker process.BlockTracker,
	genesisTime int64,
	enableEpochsHandler common.EnableEpochsHandler,
	proofsPool process.ProofsPool,
) (*metaForkDetector, error) {
	if check.IfNil(log) {
		return nil, common.ErrNilLogger
	}
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
	if check.IfNil(proofsPool) {
		return nil, process.ErrNilProofsPool
	}

	genesisHdr, _, err := blockTracker.GetSelfNotarizedHeader(core.MetachainShardId, 0)
	if err != nil {
		return nil, err
	}

	bfd := &baseForkDetector{
		log:                 log,
		roundHandler:        roundHandler,
		blackListHandler:    blackListHandler,
		genesisTime:         genesisTime,
		blockTracker:        blockTracker,
		genesisNonce:        genesisHdr.GetNonce(),
		genesisRound:        genesisHdr.GetRound(),
		genesisEpoch:        genesisHdr.GetEpoch(),
		enableEpochsHandler: enableEpochsHandler,
		proofsPool:          proofsPool,
	}

	bfd.headers = make(map[uint64][]*headerInfo)
	bfd.fork.checkpoint = make([]*checkpointInfo, 0)
	checkpoint := &checkpointInfo{
		nonce: bfd.genesisNonce,
		round: bfd.genesisRound,
	}
	bfd.setFinalCheckpoint(checkpoint)
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
	mfd.setFinalCheckpoint(mfd.lastCheckpoint())
	newCheckpoint := &checkpointInfo{nonce: header.GetNonce(), round: header.GetRound(), hash: headerHash}
	mfd.addCheckpoint(newCheckpoint)
	if mfd.enableEpochsHandler.IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, header.GetEpoch()) {
		mfd.setFinalCheckpoint(newCheckpoint)
	}
	mfd.removePastOrInvalidRecords()
}

func (mfd *metaForkDetector) computeFinalCheckpoint() {
}

// IsInterfaceNil returns true if there is no value under the interface
func (mfd *metaForkDetector) IsInterfaceNil() bool {
	return mfd == nil
}
