package sync

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// shardForkDetector implements the shard fork detector mechanism
type shardForkDetector struct {
	*baseForkDetector
}

// NewShardForkDetector method creates a new shardForkDetector object
func NewShardForkDetector(
	rounder consensus.Rounder,
	blackListHandler process.BlackListHandler,
	blockTracker process.BlockTracker,
) (*shardForkDetector, error) {

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
		blockTracker:     blockTracker,
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

	sfd.activateForcedForkIfNeeded(header, state, sharding.MetachainShardId)

	isHeaderReceivedTooLate := sfd.isHeaderReceivedTooLate(header, state, process.ShardBlockFinality)
	if isHeaderReceivedTooLate {
		state = process.BHReceivedTooLate
	}

	sfd.append(&headerInfo{
		nonce: header.GetNonce(),
		round: header.GetRound(),
		hash:  headerHash,
		state: state,
	})

	if state == process.BHProcessed {
		sfd.AddFinalHeaders(finalHeaders, finalHeadersHashes)
		sfd.addCheckpoint(&checkpointInfo{nonce: header.GetNonce(), round: header.GetRound()})
		sfd.removePastOrInvalidRecords()
	}

	probableHighestNonce := sfd.computeProbableHighestNonce()
	sfd.setLastBlockRound(uint64(sfd.rounder.Index()))
	sfd.setProbableHighestNonce(probableHighestNonce)

	return nil
}

// AddFinalHeaders method adds new final headers to headers map
func (sfd *shardForkDetector) AddFinalHeaders(finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte) {
	for i := 0; i < len(finalHeaders); i++ {
		sfd.append(&headerInfo{
			nonce: finalHeaders[i].GetNonce(),
			round: finalHeaders[i].GetRound(),
			hash:  finalHeadersHashes[i],
			state: process.BHNotarized,
		})
	}

	sfd.computeFinalCheckpoint()
}

func (sfd *shardForkDetector) computeFinalCheckpoint() {
	finalCheckPoint := sfd.finalCheckpoint()

	sfd.mutHeaders.RLock()
	for nonce, hdrInfos := range sfd.headers {
		indexBHNotarized := -1
		indexBHProcessed := -1
		for index, hdrInfo := range hdrInfos {
			if hdrInfo.state == process.BHNotarized {
				indexBHNotarized = index
			}
			if hdrInfo.state == process.BHProcessed {
				indexBHProcessed = index
			}
		}

		if indexBHNotarized != -1 && indexBHProcessed != -1 {
			if finalCheckPoint.nonce < nonce {
				if bytes.Equal(hdrInfos[indexBHNotarized].hash, hdrInfos[indexBHProcessed].hash) {
					sfd.setFinalCheckpoint(&checkpointInfo{nonce: nonce, round: hdrInfos[indexBHNotarized].round})
				}
			}
		}
	}
	sfd.mutHeaders.RUnlock()
}
