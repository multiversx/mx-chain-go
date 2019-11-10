package sync

import (
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
	}

	bfd.headers = make(map[uint64][]*headerInfo)
	checkpoint := &checkpointInfo{}
	bfd.setFinalCheckpoint(checkpoint)
	bfd.addCheckpoint(checkpoint)

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

	if header == nil || header.IsInterfaceNil() {
		return ErrNilHeader
	}
	if headerHash == nil {
		return ErrNilHash
	}

	err := sfd.checkBlockBasicValidity(header, state)
	if err != nil {
		return err
	}

	sfd.activateForcedForkIfNeeded(header, state)

	err = sfd.shouldAddBlockInForkDetector(header, state, process.ShardBlockFinality)
	if err != nil {
		return err
	}

	if state == process.BHProcessed {
		sfd.addFinalHeaders(finalHeaders, finalHeadersHashes)
		sfd.addCheckpoint(&checkpointInfo{nonce: header.GetNonce(), round: header.GetRound()})
		sfd.removePastOrInvalidRecords()
		sfd.setIsNotarizedShardStuck(isNotarizedShardStuck)
	}

	sfd.append(&headerInfo{
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
