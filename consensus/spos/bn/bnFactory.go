package bn

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"time"
)

// factory defines the data needed by this factory to create all the subrounds and give them their specific
// functionality
type factory struct {
	consensusDataContainer spos.ConsensusDataContainerInterface
	consensusState         *spos.ConsensusState
	worker                 *worker
}

// NewFactory creates a new consensusState object
func NewFactory(
	consensusDataContainer spos.ConsensusDataContainerInterface,
	consensusState *spos.ConsensusState,
	worker *worker,
) (*factory, error) {

	err := checkNewFactoryParams(
		consensusDataContainer,
		consensusState,
		worker,
	)

	if err != nil {
		return nil, err
	}

	fct := factory{
		consensusDataContainer: consensusDataContainer,
		consensusState:         consensusState,
		worker:                 worker,
	}

	return &fct, nil
}

func checkNewFactoryParams(
	container spos.ConsensusDataContainerInterface,
	state *spos.ConsensusState,
	worker *worker,
) error {
	consensusDataContainerValidator := spos.ConsensusContainerValidator{}

	err := consensusDataContainerValidator.ValidateConsensusDataContainer(container)
	if err != nil {
		return err
	}

	if state == nil {
		return spos.ErrNilConsensusState
	}

	if worker == nil {
		return spos.ErrNilWorker
	}

	return nil
}

// GenerateSubrounds will generate the subrounds used in Belare & Naveen Cns
func (fct *factory) GenerateSubrounds() error {
	fct.initConsensusThreshold()
	fct.consensusDataContainer.Chronology().RemoveAllSubrounds()
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
	return fct.consensusDataContainer.Rounder().TimeDuration()
}

func (fct *factory) generateStartRoundSubround() error {
	subround, err := NewSubround(
		-1,
		SrStartRound,
		SrBlock,
		int64(float64(fct.getTimeDuration())*srStartStartTime),
		int64(float64(fct.getTimeDuration())*srStartEndTime),
		getSubroundName(SrStartRound),
		fct.consensusState,
		fct.worker.consensusStateChangedChannels,
		fct.consensusDataContainer,
	)

	if err != nil {
		return err
	}

	subroundStartRound, err := NewSubroundStartRound(
		subround,
		fct.worker.extend,
	)

	if err != nil {
		return err
	}

	fct.consensusDataContainer.Chronology().AddSubround(subroundStartRound)

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
		fct.consensusState,
		fct.worker.consensusStateChangedChannels,
		fct.consensusDataContainer,
	)

	if err != nil {
		return err
	}

	subroundBlock, err := NewSubroundBlock(
		subround,
		fct.worker.sendConsensusMessage,
		fct.worker.extend,
	)

	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(MtBlockBody, subroundBlock.receivedBlockBody)
	fct.worker.AddReceivedMessageCall(MtBlockHeader, subroundBlock.receivedBlockHeader)
	fct.consensusDataContainer.Chronology().AddSubround(subroundBlock)

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
		fct.consensusState,
		fct.worker.consensusStateChangedChannels,
		fct.consensusDataContainer,
	)

	if err != nil {
		return err
	}

	subroundCommitmentHash, err := NewSubroundCommitmentHash(
		subround,
		fct.worker.sendConsensusMessage,
		fct.worker.extend,
	)

	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(MtCommitmentHash, subroundCommitmentHash.receivedCommitmentHash)
	fct.consensusDataContainer.Chronology().AddSubround(subroundCommitmentHash)

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
		fct.consensusState,
		fct.worker.consensusStateChangedChannels,
		fct.consensusDataContainer,
	)

	if err != nil {
		return err
	}

	subroundBitmap, err := NewSubroundBitmap(
		subround,
		fct.worker.sendConsensusMessage,
		fct.worker.extend,
	)

	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(MtBitmap, subroundBitmap.receivedBitmap)
	fct.consensusDataContainer.Chronology().AddSubround(subroundBitmap)

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
		fct.consensusState,
		fct.worker.consensusStateChangedChannels,
		fct.consensusDataContainer,
	)

	if err != nil {
		return err
	}

	subroundCommitment, err := NewSubroundCommitment(
		subround,
		fct.worker.sendConsensusMessage,
		fct.worker.extend,
	)

	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(MtCommitment, subroundCommitment.receivedCommitment)
	fct.consensusDataContainer.Chronology().AddSubround(subroundCommitment)

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
		fct.consensusState,
		fct.worker.consensusStateChangedChannels,
		fct.consensusDataContainer,
	)

	if err != nil {
		return err
	}

	subroundSignature, err := NewSubroundSignature(
		subround,
		fct.worker.sendConsensusMessage,
		fct.worker.extend,
	)

	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(MtSignature, subroundSignature.receivedSignature)
	fct.consensusDataContainer.Chronology().AddSubround(subroundSignature)

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
		fct.consensusState,
		fct.worker.consensusStateChangedChannels,
		fct.consensusDataContainer,
	)

	if err != nil {
		return err
	}

	subroundEndRound, err := NewSubroundEndRound(
		subround,
		fct.worker.BroadcastBlock,
		fct.worker.extend,
	)

	if err != nil {
		return err
	}

	fct.consensusDataContainer.Chronology().AddSubround(subroundEndRound)

	return nil
}

func (fct *factory) initConsensusThreshold() {
	pbftThreshold := fct.consensusState.ConsensusGroupSize()*2/3 + 1

	fct.consensusState.SetThreshold(SrBlock, 1)
	fct.consensusState.SetThreshold(SrCommitmentHash, pbftThreshold)
	fct.consensusState.SetThreshold(SrBitmap, pbftThreshold)
	fct.consensusState.SetThreshold(SrCommitment, pbftThreshold)
	fct.consensusState.SetThreshold(SrSignature, pbftThreshold)
}
