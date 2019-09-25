package sync

import (
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

	err = sfd.checkShardBlockValidity(header, state)
	if err != nil {
		return err
	}

	if state == process.BHProcessed {
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

func (sfd *shardForkDetector) checkShardBlockValidity(header data.HeaderHandler, state process.BlockHeaderState) error {
	noncesDifference := int64(sfd.ProbableHighestNonce()) - int64(header.GetNonce())
	isSyncing := state == process.BHReceived && noncesDifference > process.MaxNoncesDifference
	if state == process.BHProcessed || isSyncing {
		return nil
	}

	roundTooOld := int64(header.GetRound()) < sfd.rounder.Index()-process.ShardBlockFinality
	if roundTooOld {
		return ErrLowerRoundInBlock
	}

	return nil
}
