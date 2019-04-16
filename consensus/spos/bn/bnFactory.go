package bn

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"time"
)

func GetStringValue(msgType spos.MessageType) string {
	switch msgType {
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

// factory defines the data needed by this factory to create all the subrounds and give them their specific
// functionality
type factory struct {
	consensusDataContainer *spos.ConsensusDataContainer
	worker                 *spos.SPOSWorker
}

// NewFactory creates a new consensusState object
func NewFactory(
	consensusDataContainer *spos.ConsensusDataContainer,
	worker *spos.SPOSWorker,
) (*factory, error) {

	err := checkNewFactoryParams(
		consensusDataContainer,
		worker,
	)

	if err != nil {
		return nil, err
	}

	fct := factory{
		consensusDataContainer: consensusDataContainer,
		worker:                 worker,
	}

	return &fct, nil
}

func checkNewFactoryParams(
	container *spos.ConsensusDataContainer,
	worker *spos.SPOSWorker,
) error {
	err := spos.ValidateConsensusDataContainer(container)
	if err != nil {
		return err
	}

	if worker == nil {
		return spos.ErrNilWorker
	}

	return nil
}

// GenerateSubrounds will generate the subrounds used in Belare & Naveen Cns
func (fct *factory) GenerateSubrounds() error {
	fct.initConsensusThreshold()
	fct.consensusDataContainer.GetChronology().RemoveAllSubrounds()
	fct.worker.RemoveAllReceivedMessagesCalls()

	err := fct.generateStartRoundSubround()

	if err != nil {
		return err
	}

	err = fct.generateBlockSubround()

	if err != nil {
		return err
	}

	err = fct.generateCommitmentHashSubround()

	if err != nil {
		return err
	}

	err = fct.generateBitmapSubround()

	if err != nil {
		return err
	}

	err = fct.generateCommitmentSubround()

	if err != nil {
		return err
	}

	err = fct.generateSignatureSubround()

	if err != nil {
		return err
	}

	err = fct.generateEndRoundSubround()

	if err != nil {
		return err
	}

	return nil
}

func (fct *factory) getTimeDuration() time.Duration {
	return fct.consensusDataContainer.GetRounder().TimeDuration()
}

func (fct *factory) generateStartRoundSubround() error {
	subround, err := NewSubround(
		-1,
		SrStartRound,
		SrBlock,
		int64(float64(fct.getTimeDuration())*srStartStartTime),
		int64(float64(fct.getTimeDuration())*srStartEndTime),
		getSubroundName(SrStartRound),
		fct.worker.ConsensusStateChangedChannels,
		fct.consensusDataContainer,
	)

	if err != nil {
		return err
	}

	subroundStartRound, err := NewSubroundStartRound(
		subround,
		fct.worker.Extend,
	)

	if err != nil {
		return err
	}

	fct.consensusDataContainer.GetChronology().AddSubround(subroundStartRound)

	return nil
}

func (fct *factory) generateBlockSubround() error {

	subround, err := NewSubround(
		SrStartRound,
		SrBlock,
		SrCommitmentHash,
		int64(float64(fct.getTimeDuration())*srBlockStartTime),
		int64(float64(fct.getTimeDuration())*srBlockEndTime),
		getSubroundName(SrBlock),
		fct.worker.ConsensusStateChangedChannels,
		fct.consensusDataContainer,
	)

	if err != nil {
		return err
	}

	subroundBlock, err := NewSubroundBlock(
		subround,
		fct.worker.SendConsensusMessage,
		fct.worker.Extend,
	)

	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(MtBlockBody, subroundBlock.receivedBlockBody)
	fct.worker.AddReceivedMessageCall(MtBlockHeader, subroundBlock.receivedBlockHeader)
	fct.consensusDataContainer.GetChronology().AddSubround(subroundBlock)

	return nil
}

func (fct *factory) generateCommitmentHashSubround() error {
	subround, err := NewSubround(
		SrBlock,
		SrCommitmentHash,
		SrBitmap,
		int64(float64(fct.getTimeDuration())*srCommitmentHashStartTime),
		int64(float64(fct.getTimeDuration())*srCommitmentHashEndTime),
		getSubroundName(SrCommitmentHash),
		fct.worker.ConsensusStateChangedChannels,
		fct.consensusDataContainer,
	)

	if err != nil {
		return err
	}

	subroundCommitmentHash, err := NewSubroundCommitmentHash(
		subround,
		fct.worker.SendConsensusMessage,
		fct.worker.Extend,
	)

	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(MtCommitmentHash, subroundCommitmentHash.receivedCommitmentHash)
	fct.consensusDataContainer.GetChronology().AddSubround(subroundCommitmentHash)

	return nil
}

func (fct *factory) generateBitmapSubround() error {
	subround, err := NewSubround(
		SrCommitmentHash,
		SrBitmap,
		SrCommitment,
		int64(float64(fct.getTimeDuration())*srBitmapStartTime),
		int64(float64(fct.getTimeDuration())*srBitmapEndTime),
		getSubroundName(SrBitmap),
		fct.worker.ConsensusStateChangedChannels,
		fct.consensusDataContainer,
	)

	if err != nil {
		return err
	}

	subroundBitmap, err := NewSubroundBitmap(
		subround,
		fct.worker.SendConsensusMessage,
		fct.worker.Extend,
	)

	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(MtBitmap, subroundBitmap.receivedBitmap)
	fct.consensusDataContainer.GetChronology().AddSubround(subroundBitmap)

	return nil
}

func (fct *factory) generateCommitmentSubround() error {
	subround, err := NewSubround(
		SrBitmap,
		SrCommitment,
		SrSignature,
		int64(float64(fct.getTimeDuration())*srCommitmentStartTime),
		int64(float64(fct.getTimeDuration())*srCommitmentEndTime),
		getSubroundName(SrCommitment),
		fct.worker.ConsensusStateChangedChannels,
		fct.consensusDataContainer,
	)

	if err != nil {
		return err
	}

	subroundCommitment, err := NewSubroundCommitment(
		subround,
		fct.worker.SendConsensusMessage,
		fct.worker.Extend,
	)

	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(MtCommitment, subroundCommitment.receivedCommitment)
	fct.consensusDataContainer.GetChronology().AddSubround(subroundCommitment)

	return nil
}

func (fct *factory) generateSignatureSubround() error {
	subround, err := NewSubround(
		SrCommitment,
		SrSignature,
		SrEndRound,
		int64(float64(fct.getTimeDuration())*srSignatureStartTime),
		int64(float64(fct.getTimeDuration())*srSignatureEndTime),
		getSubroundName(SrSignature),
		fct.worker.ConsensusStateChangedChannels,
		fct.consensusDataContainer,
	)

	if err != nil {
		return err
	}

	subroundSignature, err := NewSubroundSignature(
		subround,
		fct.worker.SendConsensusMessage,
		fct.worker.Extend,
	)

	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(MtSignature, subroundSignature.receivedSignature)
	fct.consensusDataContainer.GetChronology().AddSubround(subroundSignature)

	return nil
}

func (fct *factory) generateEndRoundSubround() error {
	subround, err := NewSubround(
		SrSignature,
		SrEndRound,
		-1,
		int64(float64(fct.getTimeDuration())*srEndStartTime),
		int64(float64(fct.getTimeDuration())*srEndEndTime),
		getSubroundName(SrEndRound),
		fct.worker.ConsensusStateChangedChannels,
		fct.consensusDataContainer,
	)

	if err != nil {
		return err
	}

	subroundEndRound, err := NewSubroundEndRound(
		subround,
		fct.worker.BroadcastBlock,
		fct.worker.Extend,
	)

	if err != nil {
		return err
	}

	fct.consensusDataContainer.GetChronology().AddSubround(subroundEndRound)

	return nil
}

func (fct *factory) initConsensusThreshold() {
	pbftThreshold := fct.consensusDataContainer.GetConsensusState().ConsensusGroupSize()*2/3 + 1

	fct.consensusDataContainer.GetConsensusState().SetThreshold(SrBlock, 1)
	fct.consensusDataContainer.GetConsensusState().SetThreshold(SrCommitmentHash, pbftThreshold)
	fct.consensusDataContainer.GetConsensusState().SetThreshold(SrBitmap, pbftThreshold)
	fct.consensusDataContainer.GetConsensusState().SetThreshold(SrCommitment, pbftThreshold)
	fct.consensusDataContainer.GetConsensusState().SetThreshold(SrSignature, pbftThreshold)
}
