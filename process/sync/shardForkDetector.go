package sync

import (
	"bytes"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"math"
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
	genesisTime int64,
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
		genesisTime:      genesisTime,
		blockTracker:     blockTracker,
	}

	bfd.headers = make(map[uint64][]*headerInfo)
	checkpoint := &checkpointInfo{}
	bfd.setFinalCheckpoint(checkpoint)
	bfd.addCheckpoint(checkpoint)
	bfd.fork.nonce = math.MaxUint64

	sfd := shardForkDetector{
		baseForkDetector: bfd,
	}

	sfd.blockTracker.RegisterSelfNotarizedHeadersHandler(sfd.addSelfNotarizedHeaders)

	return &sfd, nil
}

// AddHeader method adds a new header to headers map
func (sfd *shardForkDetector) AddHeader(
	header data.HeaderHandler,
	headerHash []byte,
	state process.BlockHeaderState,
	selfNotarizedHeaders []data.HeaderHandler,
	selfNotarizedHeadersHashes [][]byte,
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

	isHeaderReceivedTooLate := sfd.isHeaderReceivedTooLate(header, state, process.BlockFinality)
	if isHeaderReceivedTooLate {
		state = process.BHReceivedTooLate
	}

	_ = sfd.append(&headerInfo{
		epoch: header.GetEpoch(),
		nonce: header.GetNonce(),
		round: header.GetRound(),
		hash:  headerHash,
		state: state,
	})

	if state == process.BHProcessed {
		sfd.addSelfNotarizedHeaders(sharding.MetachainShardId, selfNotarizedHeaders, selfNotarizedHeadersHashes)
		sfd.addCheckpoint(&checkpointInfo{nonce: header.GetNonce(), round: header.GetRound(), hash: headerHash})
		sfd.removePastOrInvalidRecords()
	}

	probableHighestNonce := sfd.computeProbableHighestNonce()
	sfd.setLastBlockRound(uint64(sfd.rounder.Index()))
	sfd.setProbableHighestNonce(probableHighestNonce)

	return nil
}

func (sfd *shardForkDetector) addSelfNotarizedHeaders(
	shardID uint32,
	selfNotarizedHeaders []data.HeaderHandler,
	selfNotarizedHeadersHashes [][]byte,
) {
	if shardID != sharding.MetachainShardId || len(selfNotarizedHeaders) == 0 {
		return
	}

	selfNotarizedHeaderAdded := false
	finalNonce := sfd.finalCheckpoint().nonce

	for i := 0; i < len(selfNotarizedHeaders); i++ {
		if selfNotarizedHeaders[i].GetNonce() <= finalNonce {
			continue
		}

		if sfd.append(&headerInfo{
			nonce: selfNotarizedHeaders[i].GetNonce(),
			round: selfNotarizedHeaders[i].GetRound(),
			hash:  selfNotarizedHeadersHashes[i],
			state: process.BHNotarized,
		}) {
			log.Debug("added self notarized header",
				"shard", selfNotarizedHeaders[i].GetShardID(),
				"round", selfNotarizedHeaders[i].GetRound(),
				"nonce", selfNotarizedHeaders[i].GetNonce(),
				"hash", selfNotarizedHeadersHashes[i])

			selfNotarizedHeaderAdded = true
		}
	}

	if selfNotarizedHeaderAdded {
		sfd.computeFinalCheckpoint()
	}
}

func (sfd *shardForkDetector) computeFinalCheckpoint() {
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
			finalNonce := sfd.finalCheckpoint().nonce
			if finalNonce < nonce {
				if bytes.Equal(hdrInfos[indexBHNotarized].hash, hdrInfos[indexBHProcessed].hash) {
					sfd.setFinalCheckpoint(&checkpointInfo{nonce: nonce, round: hdrInfos[indexBHNotarized].round, hash: hdrInfos[indexBHNotarized].hash})
				}
			}
		}
	}
	sfd.mutHeaders.RUnlock()
}
