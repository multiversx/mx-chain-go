package sync

import (
	"math"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

// metaForkDetector implements the meta fork detector mechanism
type metaForkDetector struct {
	*baseForkDetector
}

// NewMetaForkDetector method creates a new metaForkDetector object
func NewMetaForkDetector(
	rounder consensus.Rounder,
	blackListHandler process.BlackListHandler,
	blockTracker process.BlockTracker,
	genesisTime int64,
) (*metaForkDetector, error) {

	if check.IfNil(rounder) {
		return nil, process.ErrNilRounder
	}
	if check.IfNil(blackListHandler) {
		return nil, process.ErrNilBlackListHandler
	}
	if check.IfNil(blockTracker) {
		return nil, process.ErrNilBlockTracker
	}

	bfd := &baseForkDetector{
		rounder:          rounder,
		blackListHandler: blackListHandler,
		genesisTime:      genesisTime,
		blockTracker:     blockTracker,
	}

	bfd.headers = make(map[uint64][]*headerInfo)
	bfd.fork.checkpoint = make([]*checkpointInfo, 0)
	checkpoint := &checkpointInfo{}
	bfd.setFinalCheckpoint(checkpoint)
	bfd.addCheckpoint(checkpoint)
	bfd.fork.rollBackNonce = math.MaxUint64

	mfd := metaForkDetector{
		baseForkDetector: bfd,
	}

	return &mfd, nil
}

// AddHeader method adds a new header to headers map
func (mfd *metaForkDetector) AddHeader(
	header data.HeaderHandler,
	headerHash []byte,
	state process.BlockHeaderState,
	_ []data.HeaderHandler,
	_ [][]byte,
) error {

	if check.IfNil(header) {
		return ErrNilHeader
	}
	if headerHash == nil {
		return ErrNilHash
	}

	err := mfd.checkBlockBasicValidity(header, headerHash, state)
	if err != nil {
		return err
	}

	if header.GetNonce() > mfd.highestNonceReceived() {
		mfd.setHighestNonceReceived(header.GetNonce())
		log.Debug("forkDetector.AddHeader.setHighestNonceReceived",
			"highest nonce received", mfd.highestNonceReceived())
	}

	if state == process.BHProposed {
		return nil
	}

	isHeaderReceivedTooLate := mfd.isHeaderReceivedTooLate(header, state, process.BlockFinality)
	if isHeaderReceivedTooLate {
		state = process.BHReceivedTooLate
	}

	appended := mfd.append(&headerInfo{
		epoch: header.GetEpoch(),
		nonce: header.GetNonce(),
		round: header.GetRound(),
		hash:  headerHash,
		state: state,
	})
	if !appended {
		return nil
	}

	if state == process.BHProcessed {
		mfd.setFinalCheckpoint(mfd.lastCheckpoint())
		mfd.addCheckpoint(&checkpointInfo{nonce: header.GetNonce(), round: header.GetRound(), hash: headerHash})
		mfd.removePastOrInvalidRecords()
	}

	probableHighestNonce := mfd.computeProbableHighestNonce()
	mfd.setProbableHighestNonce(probableHighestNonce)

	log.Debug("forkDetector.AddHeader",
		"round", header.GetRound(),
		"nonce", header.GetNonce(),
		"hash", headerHash,
		"state", state,
		"probable highest nonce", mfd.probableHighestNonce(),
		"last check point nonce", mfd.lastCheckpoint().nonce,
		"final check point nonce", mfd.finalCheckpoint().nonce)

	return nil
}
