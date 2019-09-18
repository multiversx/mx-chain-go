package sync

import (
	"bytes"
	"math"
	"strings"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

// metaForkDetector implements the meta fork detector mechanism
type metaForkDetector struct {
	*baseForkDetector
}

// NewMetaForkDetector method creates a new metaForkDetector object
func NewMetaForkDetector(rounder consensus.Rounder) (*metaForkDetector, error) {
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
	finalHeaders []data.HeaderHandler,
	finalHeadersHashes [][]byte,
) error {

	if header == nil || header.IsInterfaceNil() {
		return ErrNilHeader
	}
	if headerHash == nil {
		return ErrNilHash
	}

	err := mfd.checkBlockBasicValidity(header, state)
	if err != nil {
		return err
	}

	err = mfd.checkMetaBlockValidity(header)
	if err != nil {
		return err
	}

	if state == process.BHProcessed {
		// add check point and remove all the past checkpoints/headers
		mfd.setFinalCheckpoint(mfd.lastCheckpoint())
		mfd.addCheckpoint(&checkpointInfo{nonce: header.GetNonce(), round: header.GetRound()})
		mfd.removePastOrInvalidRecords()
	}

	mfd.append(&headerInfo{
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

func (mfd *metaForkDetector) checkMetaBlockValidity(header data.HeaderHandler) error {
	roundTooOld := int64(header.GetRound()) < mfd.rounder.Index()-process.MetaBlockFinality
	if roundTooOld {
		return ErrLowerRoundInBlock
	}

	return nil
}

// CheckFork method checks if the node could be on the fork
func (mfd *metaForkDetector) CheckFork() (bool, uint64, []byte) {
	var lowestForkNonce uint64
	var hashOfLowestForkNonce []byte
	var lowestRoundInForkNonce uint64
	var forkHeaderHash []byte
	var selfHdrInfo *headerInfo
	lowestForkNonce = math.MaxUint64
	hashOfLowestForkNonce = nil
	forkDetected := false

	mfd.mutHeaders.Lock()
	for nonce, hdrInfos := range mfd.headers {
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

			if hdrInfos[i].round < lowestRoundInForkNonce {
				lowestRoundInForkNonce = hdrInfos[i].round
				forkHeaderHash = hdrInfos[i].hash
				continue
			}

			if hdrInfos[i].round == lowestRoundInForkNonce && bytes.Compare(hdrInfos[i].hash, forkHeaderHash) < 0 {
				forkHeaderHash = hdrInfos[i].hash
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
			delete(mfd.headers, nonce)
			mfd.headers[nonce] = []*headerInfo{selfHdrInfo}
			continue
		}

		forkDetected = true
		if nonce < lowestForkNonce {
			lowestForkNonce = nonce
			hashOfLowestForkNonce = forkHeaderHash
		}
	}
	mfd.mutHeaders.Unlock()

	return forkDetected, lowestForkNonce, hashOfLowestForkNonce
}
