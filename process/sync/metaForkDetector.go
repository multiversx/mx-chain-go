package sync

import (
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
	blackList process.BlackListHandler,
) (*metaForkDetector, error) {

	if check.IfNil(rounder) {
		return nil, process.ErrNilRounder
	}
	if check.IfNil(blackList) {
		return nil, process.ErrNilBlackListHandler
	}

	bfd := &baseForkDetector{
		rounder:   rounder,
		blackList: blackList,
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

	err = mfd.shouldAddBlockInForkDetector(header, state, process.MetaBlockFinality)
	if err != nil {
		return err
	}

	if state == process.BHProcessed {
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
