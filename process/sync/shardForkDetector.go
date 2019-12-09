package sync

import (
	"math"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

// shardForkDetector implements the shard fork detector mechanism
type shardForkDetector struct {
	*baseForkDetector
}

// NewShardForkDetector method creates a new shardForkDetector object
func NewShardForkDetector(
	rounder consensus.Rounder,
	blackListHandler process.BlackListHandler,
	genesisTime int64,
) (*shardForkDetector, error) {

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

	sfd := shardForkDetector{
		baseForkDetector: bfd,
	}

	return &sfd, nil
}

// AddHeader method adds a new header to headers map
func (sfd *shardForkDetector) AddHeader(
	header data.HeaderHandler,
	headerHash []byte,
	state process.BlockHeaderState,
	finalHeaders []data.HeaderHandler,
	finalHeadersHashes [][]byte,
	isNotarizedShardStuck bool,
) error {

	if check.IfNil(header) {
		return ErrNilHeader
	}
	if headerHash == nil {
		return ErrNilHash
	}

	err := sfd.checkBlockBasicValidity(header, headerHash, state)
	if err != nil {
		return err
	}

	sfd.activateForcedForkIfNeeded(header, state)

	isHeaderReceivedTooLate := sfd.isHeaderReceivedTooLate(header, state, process.ShardBlockFinality)
	if isHeaderReceivedTooLate {
		state = process.BHReceivedTooLate
	}

	if state == process.BHProcessed {
		sfd.addFinalHeaders(finalHeaders, finalHeadersHashes)
		sfd.addCheckpoint(&checkpointInfo{nonce: header.GetNonce(), round: header.GetRound()})
		sfd.removePastOrInvalidRecords()
		sfd.setIsNotarizedShardStuck(isNotarizedShardStuck)
	}

	sfd.append(&headerInfo{
		epoch: header.GetEpoch(),
		nonce: header.GetNonce(),
		round: header.GetRound(),
		hash:  headerHash,
		state: state,
	})

	probableHighestNonce := sfd.computeProbableHighestNonce()
	sfd.setLastBlockRound(uint64(sfd.rounder.Index()))
	sfd.setProbableHighestNonce(probableHighestNonce)

	return nil
}

func (sfd *shardForkDetector) addFinalHeaders(finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte) {
	finalCheckpointWasSet := false
	for i := 0; i < len(finalHeaders); i++ {
		isFinalHeaderNonceNotLowerThanCurrent := finalHeaders[i].GetNonce() >= sfd.finalCheckpoint().nonce
		if isFinalHeaderNonceNotLowerThanCurrent {
			if !finalCheckpointWasSet {
				sfd.setFinalCheckpoint(&checkpointInfo{nonce: finalHeaders[i].GetNonce(), round: finalHeaders[i].GetRound()})
				finalCheckpointWasSet = true
			}

			sfd.append(&headerInfo{
				nonce: finalHeaders[i].GetNonce(),
				round: finalHeaders[i].GetRound(),
				hash:  finalHeadersHashes[i],
				state: process.BHNotarized,
			})
		}
	}
}
