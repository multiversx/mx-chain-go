package sync

import (
	"math"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
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

	genesisHdr, _, err := blockTracker.GetSelfNotarizedHeader(core.MetachainShardId, 0)
	if err != nil {
		return nil, err
	}

	bfd := &baseForkDetector{
		roundHandler:     roundHandler,
		blackListHandler: blackListHandler,
		genesisTime:      genesisTime,
		blockTracker:     blockTracker,
		genesisNonce:     genesisHdr.GetNonce(),
		genesisRound:     genesisHdr.GetRound(),
		genesisEpoch:     genesisHdr.GetEpoch(),
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
	mfd.addCheckpoint(&checkpointInfo{nonce: header.GetNonce(), round: header.GetRound(), hash: headerHash})
	mfd.removePastOrInvalidRecords()
}

func (mfd *metaForkDetector) computeFinalCheckpoint() {
}
