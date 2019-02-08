package bn

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
)

const (
	// SrStartRound defines ID of subround "Start round"
	SrStartRound chronology.SubroundId = iota
	// SrBlock defines ID of subround "block"
	SrBlock
	// SrCommitmentHash defines ID of subround "commitment hash"
	SrCommitmentHash
	// SrBitmap defines ID of subround "bitmap"
	SrBitmap
	// SrCommitment defines ID of subround "commitment"
	SrCommitment
	// SrSignature defines ID of subround "signature"
	SrSignature
	// SrEndRound defines ID of subround "End round"
	SrEndRound
	// SrAdvance defines ID of subround "Advance"
	SrAdvance
)

//TODO: maximum transactions in one block (this should be injected, and this const should be removed later)
const maxTransactionsInBlock = 15000

// consensusSubrounds specifies how many subrounds of consensus are in this implementation
const consensusSubrounds = 6

// maxBlockProcessingTimePercent specifies which is the max allocated time percent,
// for processing block, from the total time of one round
const maxBlockProcessingTimePercent = float64(0.85)

// MessageType specifies what type of message was received
type MessageType int

const (
	// MtUnknown defines ID of a message that has unknown Data inside
	MtUnknown MessageType = iota
	// MtBlockBody defines ID of a message that has a block body inside
	MtBlockBody
	// MtBlockHeader defines ID of a message that has a block header inside
	MtBlockHeader
	// MtCommitmentHash defines ID of a message that has a commitment hash inside
	MtCommitmentHash
	// MtBitmap defines ID of a message that has a bitmap inside
	MtBitmap
	// MtCommitment defines ID of a message that has a commitment inside
	MtCommitment
	// MtSignature defines ID of a message that has a Signature inside
	MtSignature
)

// bnFactory defines the data needed by this factory to create all the subrounds and give them their specific
// functionality
type bnFactory struct {
	wrk *Worker
}

// NewbnFactory creates a new Spos object
func NewbnFactory(
	wrk *Worker,
) (*bnFactory, error) {

	err := checkNewbnFactoryParams(
		wrk,
	)

	if err != nil {
		return nil, err
	}

	bnf := bnFactory{
		wrk: wrk,
	}

	return &bnf, nil
}

func checkNewbnFactoryParams(
	wrk *Worker,
) error {
	if wrk == nil {
		return spos.ErrNilWorker
	}

	return nil
}

// GenerateSubrounds will generate the subrounds used in Belare & Naveen SPoS
func (bnf *bnFactory) GenerateSubrounds() {
	chr := bnf.wrk.SPoS.Chr
	sps := bnf.wrk.SPoS

	pbftThreshold := sps.ConsensusGroupSize()*2/3 + 1

	sps.SetThreshold(SrBlock, 1)
	sps.SetThreshold(SrCommitmentHash, pbftThreshold)
	sps.SetThreshold(SrBitmap, pbftThreshold)
	sps.SetThreshold(SrCommitment, pbftThreshold)
	sps.SetThreshold(SrSignature, pbftThreshold)

	chr.RemoveAllSubrounds()

	roundDuration := chr.Round().TimeDuration()

	chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(SrStartRound),
		chronology.SubroundId(SrBlock), int64(roundDuration*5/100),
		getSubroundName(SrStartRound),
		bnf.wrk.doStartRoundJob,
		bnf.wrk.extendStartRound,
		bnf.wrk.checkStartRoundConsensus))

	chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(SrBlock),
		chronology.SubroundId(SrCommitmentHash),
		int64(roundDuration*25/100),
		getSubroundName(SrBlock),
		bnf.wrk.doBlockJob,
		bnf.wrk.extendBlock,
		bnf.wrk.checkBlockConsensus))

	chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(SrCommitmentHash),
		chronology.SubroundId(SrBitmap),
		int64(roundDuration*40/100),
		getSubroundName(SrCommitmentHash),
		bnf.wrk.doCommitmentHashJob,
		bnf.wrk.extendCommitmentHash,
		bnf.wrk.checkCommitmentHashConsensus))

	chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(SrBitmap),
		chronology.SubroundId(SrCommitment),
		int64(roundDuration*55/100),
		getSubroundName(SrBitmap),
		bnf.wrk.doBitmapJob,
		bnf.wrk.extendBitmap,
		bnf.wrk.checkBitmapConsensus))

	chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(SrCommitment),
		chronology.SubroundId(SrSignature),
		int64(roundDuration*70/100),
		getSubroundName(SrCommitment),
		bnf.wrk.doCommitmentJob,
		bnf.wrk.extendCommitment,
		bnf.wrk.checkCommitmentConsensus))

	chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(SrSignature),
		chronology.SubroundId(SrEndRound),
		int64(roundDuration*85/100),
		getSubroundName(SrSignature),
		bnf.wrk.doSignatureJob,
		bnf.wrk.extendSignature,
		bnf.wrk.checkSignatureConsensus))

	chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(SrEndRound),
		chronology.SubroundId(SrAdvance),
		int64(roundDuration*95/100),
		getSubroundName(SrEndRound),
		bnf.wrk.doEndRoundJob,
		bnf.wrk.extendEndRound,
		bnf.wrk.checkEndRoundConsensus))

	chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(SrAdvance),
		-1,
		int64(roundDuration*100/100),
		getSubroundName(SrAdvance),
		bnf.wrk.doAdvanceJob,
		nil,
		bnf.wrk.checkAdvanceConsensus))
}

// getMessageTypeName method returns the name of the message from a given message ID
func getMessageTypeName(messageType MessageType) string {
	switch messageType {
	case MtBlockBody:
		return "(BLOCK_BODY)"
	case MtBlockHeader:
		return "(BLOCK_HEADER)"
	case MtCommitmentHash:
		return "(COMMITMENT_HASH)"
	case MtBitmap:
		return "(BITMAP)"
	case MtCommitment:
		return "(COMMITMENT)"
	case MtSignature:
		return "(SIGNATURE)"
	case MtUnknown:
		return "(UNKNOWN)"
	default:
		return "Undefined message type"
	}
}

// getSubroundName returns the name of each subround from a given subround ID
func getSubroundName(subroundId chronology.SubroundId) string {
	switch subroundId {
	case SrStartRound:
		return "(START_ROUND)"
	case SrBlock:
		return "(BLOCK)"
	case SrCommitmentHash:
		return "(COMMITMENT_HASH)"
	case SrBitmap:
		return "(BITMAP)"
	case SrCommitment:
		return "(COMMITMENT)"
	case SrSignature:
		return "(SIGNATURE)"
	case SrEndRound:
		return "(END_ROUND)"
	case SrAdvance:
		return "(ADVANCE)"
	default:
		return "Undefined subround"
	}
}
