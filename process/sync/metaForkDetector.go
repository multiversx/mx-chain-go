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
	genesisTime int64,
) (*metaForkDetector, error) {

	if check.IfNil(rounder) {
		return nil, process.ErrNilRounder
	}
	if check.IfNil(blackListHandler) {
		return nil, process.ErrNilBlackListHandler
	}

	bfd := &baseForkDetector{
		rounder:          rounder,
		blackListHandler: blackListHandler,
		genesisTime:      genesisTime,
	}

	bfd.headers = make(map[uint64][]*headerInfo)
	checkpoint := &checkpointInfo{}
	bfd.setFinalCheckpoint(checkpoint)
	bfd.addCheckpoint(checkpoint)
	bfd.fork.nonce = math.MaxUint64

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
	isNotarizedShardStuck bool,
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

	mfd.activateForcedForkOnConsensusStuckIfNeeded(header, state)

	isHeaderReceivedTooLate := mfd.isHeaderReceivedTooLate(header, state, process.MetaBlockFinality)
	if isHeaderReceivedTooLate {
		state = process.BHReceivedTooLate
	}

	if state == process.BHProcessed {
		mfd.setFinalCheckpoint(mfd.lastCheckpoint())
		mfd.addCheckpoint(&checkpointInfo{nonce: header.GetNonce(), round: header.GetRound()})
		mfd.removePastOrInvalidRecords()
		mfd.setIsNotarizedShardStuck(isNotarizedShardStuck)
	}

	mfd.append(&headerInfo{
		epoch: header.GetEpoch(),
		nonce: header.GetNonce(),
		round: header.GetRound(),
		hash:  headerHash,
		state: state,
	})

	probableHighestNonce := mfd.computeProbableHighestNonce()
	mfd.setLastBlockRound(uint64(mfd.rounder.Index()))
	mfd.setProbableHighestNonce(probableHighestNonce)

	return nil
}
