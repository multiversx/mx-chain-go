package sync

import (
	"bytes"
	"math"
	"strings"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

// shardForkDetector implements the shard fork detector mechanism
type shardForkDetector struct {
	*baseForkDetector
}

// NewShardForkDetector method creates a new shardForkDetector object
func NewShardForkDetector(rounder consensus.Rounder) (*shardForkDetector, error) {
	if rounder == nil || rounder.IsInterfaceNil() {
		return nil, process.ErrNilRounder
	}

	bfd := &baseForkDetector{
		rounder: rounder,
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

	if state == process.BHProcessed {
		// add final header(s), add check point and remove all the past checkpoints/headers
		sfd.addFinalHeaders(finalHeaders, finalHeadersHashes)
		sfd.addCheckpoint(&checkpointInfo{nonce: header.GetNonce(), round: header.GetRound()})
		sfd.removePastOrInvalidRecords()
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
		isFinalHeaderNonceHigherThanCurrent := finalHeaders[i].GetNonce() > sfd.GetHighestFinalBlockNonce()
		if isFinalHeaderNonceHigherThanCurrent {
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

// CheckFork method checks if the node could be on the fork
func (sfd *shardForkDetector) CheckFork() (bool, uint64, []byte) {
	var lowestForkNonce uint64
	var hashOfLowestForkNonce []byte
	var lowestRoundInForkNonce uint64
	var forkHeaderHash []byte
	var selfHdrInfo *headerInfo
	lowestForkNonce = math.MaxUint64
	hashOfLowestForkNonce = nil
	forkDetected := false

	sfd.mutHeaders.Lock()
	for nonce, hdrInfos := range sfd.headers {
		if len(hdrInfos) == 1 {
			continue
		}

		selfHdrInfo = nil
		lowestRoundInForkNonce = math.MaxUint64
		forkHeaderHash = nil

		for i := 0; i < len(hdrInfos); i++ {
			// Proposed blocks received do not count for fork choice, as they are not valid until the consensus
			// is achieved. They should be received afterwards through sync mechanism.
			if hdrInfos[i].state == process.BHProposed {
				continue
			}

			if hdrInfos[i].state == process.BHProcessed {
				selfHdrInfo = hdrInfos[i]
				continue
			}

			if hdrInfos[i].state == process.BHNotarized {
				if lowestRoundInForkNonce > 0 {
					lowestRoundInForkNonce = 0
					forkHeaderHash = hdrInfos[i].hash
					continue
				}

				if lowestRoundInForkNonce == 0 && bytes.Compare(hdrInfos[i].hash, forkHeaderHash) < 0 {
					forkHeaderHash = hdrInfos[i].hash
				}

				continue
			}
		}

		if selfHdrInfo == nil {
			// if current nonce has not been processed yet, then skip and check the next one.
			continue
		}

		shouldSignalFork := selfHdrInfo.round > lowestRoundInForkNonce ||
			(selfHdrInfo.round == lowestRoundInForkNonce && strings.Compare(string(selfHdrInfo.hash), string(forkHeaderHash)) > 0)

		if !shouldSignalFork {
			// keep it clean so next time this position will be processed faster
			delete(sfd.headers, nonce)
			sfd.headers[nonce] = []*headerInfo{selfHdrInfo}
			continue
		}

		forkDetected = true
		if nonce < lowestForkNonce {
			lowestForkNonce = nonce
			hashOfLowestForkNonce = forkHeaderHash
		}
	}
	sfd.mutHeaders.Unlock()

	return forkDetected, lowestForkNonce, hashOfLowestForkNonce
}
